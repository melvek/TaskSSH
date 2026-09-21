package com.mestrap.core;

import com.mestrap.action.TaskAction;
import com.mestrap.entity.HostVars;
import com.mestrap.entity.Step;
import com.mestrap.entity.Task;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.LogPrinter;
import com.mestrap.utils.VariableReplacer;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * @author melvek
 */
public class TaskExecutor {

    private final ActionRegistry actionRegistry;

    public TaskExecutor(ActionRegistry actionRegistry) {
        this.actionRegistry = actionRegistry;
    }

    /**
     * 对一批主机执行一条流程
     */
    public int executeAll(Task task,
                          Map<String, HostVars> hosts,
                          Map<String, Object> globalVars,
                          Map<String, Object> cliVars) {

        int success = 0, failed = 0;

        LogPrinter.section("Start");
        LogPrinter.info("Task: " + task.getName() + " (" + task.getSteps().size() + " steps)");
        LogPrinter.info("Targets: " + hosts.size());

        int idx = 0;
        for (Map.Entry<String, HostVars> e : hosts.entrySet()) {
            idx++;
            String hostName = e.getKey();
            HostVars hostVars = e.getValue();

            LogPrinter.progress(idx, hosts.size(), "Processing: " + hostName + "[" + hostVars.getHost() + "]");

            try {
                executeOnHost(task, hostVars, globalVars, cliVars);
                success++;
                LogPrinter.success(hostName + "[" + hostVars.getHost() + "]" + " task execute completed!");
            } catch (TaskException ex) {
                failed++;
                if (ex.getStepName() != null) {
                    LogPrinter.error(hostName + " FAIL at step '"
                            + ex.getStepName() + "': " + ex.getMessage());
                } else {
                    LogPrinter.error(hostName + " FAIL: " + ex.getMessage());
                }
            } catch (Exception ex) {
                failed++;
                LogPrinter.error(hostName + " FAIL: " + ex.getMessage());
            }
        }

        LogPrinter.summary(success + failed, success, failed);
        return failed == 0 ? 0 : 1;
    }

    /**
     * 对单台主机执行整条流程
     */
    public void executeOnHost(Task task,
                              HostVars hostVars,
                              Map<String, Object> globalVars,
                              Map<String, Object> cliVars) {

        String hostName = hostVars.getHost();

        // 1. 变量池：global -> host，不含 CLI
        Map<String, Object> vars = new HashMap<>(16);
        if (globalVars != null) {
            vars.putAll(globalVars);
        }
        if (hostVars.getExtraFields() != null) {
            vars.putAll(hostVars.getExtraFields());
        }

        // 2. 解析 CLI 值里的 ${...}，每台主机各展开一次
        Map<String, Object> resolvedCli = resolveCliVars(cliVars, vars);

        List<Step> steps = task.getSteps();
        int total = steps.size();

        for (int i = 0; i < total; i++) {
            Step step = steps.get(i);

            LogPrinter.emptyLine();
            LogPrinter.info("[STEP " + (i + 1) + "/" + total + "] " + step.getName());

            TaskAction action = actionRegistry.get(step.getAction());
            if (action == null) {
                throw new TaskException(hostName, step.getName(),
                        "Unknown action: " + step.getAction(), null);
            }

            // 3. 构造 effectiveWith：step.with 基础上，CLI 覆盖（CLI 优先级最高）
            Map<String, Object> effectiveWith = new HashMap<>(8);
            if (step.getWith() != null) {
                effectiveWith.putAll(step.getWith());
            }
            if (resolvedCli != null) {
                effectiveWith.putAll(resolvedCli);
            }

            // 4. 执行
            try {
                action.execute(new ActionContext(hostVars, effectiveWith, vars));
            } catch (TaskException e) {
                throw new TaskException(hostName, step.getName(), e.getMessage(), e);
            } catch (Exception e) {
                throw new TaskException(hostName, step.getName(),
                        "Step failed: " + step.getName() + " - " + e.getMessage(), e);
            }

            // 5. 执行成功后等待
            if (step.getDelay() > 0) {
                LogPrinter.emptyLine();
                LogPrinter.info("Waiting " + step.getDelay() + "s before next step...");
                try {
                    Thread.sleep(step.getDelay() * 1000L);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new TaskException(hostName, step.getName(),
                            "Interrupted while waiting after step", e);
                }
            }
        }
    }

    /**
     * 用变量池解析 CLI 值中的 ${...}。
     * CLI 值本身不进变量池，只作为参数覆盖 step.with。
     */
    private Map<String, Object> resolveCliVars(Map<String, Object> cliVars,
                                               Map<String, Object> vars) {
        if (cliVars == null || cliVars.isEmpty()) {
            return null;
        }

        Map<String, Object> resolved = new HashMap<>(16);
        for (Map.Entry<String, Object> e : cliVars.entrySet()) {
            String raw = String.valueOf(e.getValue());
            resolved.put(e.getKey(), VariableReplacer.replace(raw, vars));
        }
        return resolved;
    }
}