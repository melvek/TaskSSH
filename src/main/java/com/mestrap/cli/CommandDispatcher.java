package com.mestrap.cli;

import com.mestrap.action.TaskAction;
import com.mestrap.core.ActionRegistry;
import com.mestrap.core.HostResolver;
import com.mestrap.core.InventoryLoader;
import com.mestrap.core.TaskExecutor;
import com.mestrap.core.TaskRegistry;
import com.mestrap.entity.HostVars;
import com.mestrap.entity.Inventory;
import com.mestrap.entity.Task;
import com.mestrap.utils.ConfirmUtil;
import com.mestrap.utils.Constant;
import com.mestrap.utils.LogPrinter;
import org.apache.commons.cli.CommandLine;
import org.apache.commons.cli.DefaultParser;
import org.apache.commons.cli.Option;
import org.apache.commons.cli.Options;
import org.apache.commons.cli.ParseException;

import java.util.*;

/**
 * 命令调度器：解析参数，加载清单，执行任务。
 *
 * @author melvek
 */
public class CommandDispatcher {

    public static final String CMD_ENCRYPT = "encrypt";
    public static final String CMD_DECRYPT = "decrypt";

    private final TaskRegistry taskRegistry = new TaskRegistry();
    private final ActionRegistry actionRegistry = new ActionRegistry();

    public int dispatch(String[] args) {

        // 无参数：打印帮助
        if (args == null || args.length == 0) {
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }

        String first = args[0];
        if (CMD_ENCRYPT.equals(first) || CMD_DECRYPT.equals(first)) {
            String[] remaining = Arrays.copyOfRange(args, 1, args.length);
            return EncryptCommand.run(first, remaining);
        }

        // ---- 1. 一次性注册所有选项 ----
        Options options = buildAllOptions();

        // ---- 2. 直接解析全部参数 ----
        CommandLine cl;
        try {
            cl = new DefaultParser().parse(options, args);
        } catch (ParseException e) {
            LogPrinter.error("Invalid arguments: " + e.getMessage());
            LogPrinter.hint("Use -h to see help");
            return 1;
        }

        // ---- 3. 元信息 ----
        if (cl.hasOption(GlobalOptions.VERSION)) {
            ShowHelp.printVersion();
            return 0;
        }
        if (cl.hasOption(GlobalOptions.HELP)) {
            ShowHelp.printGlobal(actionRegistry);
            return 0;
        }

        // ---- 4. 加载清单 ----
        String invFile = cl.hasOption(GlobalOptions.INVENTORY)
                ? cl.getOptionValue(GlobalOptions.INVENTORY)
                : Constant.DEFAULT_INVENTORY;
        Inventory inventory = InventoryLoader.load(invFile);

        // 注册用户流程（同名覆盖内置流程）
        taskRegistry.loadTasks(inventory.getTasks());

        // ---- 5. 第一个位置参数 = 流程名 ----
        List<String> positional = cl.getArgList();
        if (positional.isEmpty()) {
            LogPrinter.error("No task specified");
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }

        String taskName = positional.get(0);
        List<String> hostNames = positional.subList(1, positional.size());

        Task task = taskRegistry.resolve(taskName);
        if (task == null) {
            LogPrinter.error("Unknown task: <" + taskName + ">");
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }

        // ---- 6. 解析目标主机 ----
        Map<String, HostVars> hosts = HostResolver.resolve(hostNames, inventory, cl);

        if (hosts.isEmpty()) {
            LogPrinter.error("No target hosts specified");
            LogPrinter.hint("Usage: taskssh " + taskName + " <hosts...> [options]");
            return 1;
        }

        // ---- 7. 展示主机列表 ----
        LogPrinter.section("Target hosts");
        hosts.forEach((name, vars) -> LogPrinter.listItem(name, vars.getHost()));

        if (cl.hasOption(GlobalOptions.LIST)) {
            return 0;
        }

        // ---- 8. 确认 ----
        if (!cl.hasOption(GlobalOptions.YES)) {
            if (!ConfirmUtil.confirm("Confirm to proceed")) {
                return 0;
            }
        }

        // ---- 9. 提取 CLI 变量注入 ----
        Map<String, Object> cliVars = extractCliVars(cl);

        // ---- 10. 执行 ----
        TaskExecutor executor = new TaskExecutor(actionRegistry);
        return executor.executeAll(
                task,
                hosts,
                inventory.getGlobalVars() != null
                        ? inventory.getGlobalVars().getExtraFields()
                        : Collections.<String, Object>emptyMap(),
                cliVars
        );
    }

    /**
     * 构建选项：全局选项 + 所有 action 的选项。
     */
    private Options buildAllOptions() {
        Options options = new Options();

        for (Option o : GlobalOptions.all()) {
            options.addOption(o);
        }

        for (TaskAction action : actionRegistry.all()) {
            for (Option o : action.cliOptions()) {
                options.addOption(o);
            }
        }

        return options;
    }

    /**
     * 提取 CLI 变量注入。
     * 根据 action 声明的 cliVarMapping，把长选项值映射为参数键。
     */
    private Map<String, Object> extractCliVars(CommandLine cl) {
        Map<String, Object> vars = new HashMap<>(8);

        for (TaskAction action : actionRegistry.all()) {
            for (Map.Entry<String, String> e : action.cliVarMapping().entrySet()) {
                String cliOpt = e.getKey();
                String varKey = e.getValue();
                Option opt = findOption(action, cliOpt);
                if (opt == null) {
                    continue;
                }

                if (!cl.hasOption(cliOpt)) {
                    continue;
                }

                if (opt.hasArg()) {
                    vars.put(varKey, cl.getOptionValue(cliOpt));
                } else {
                    vars.put(varKey, "true");
                }
            }
        }
        return vars;
    }

    private Option findOption(TaskAction action, String longOpt) {
        for (Option o : action.cliOptions()) {
            if (longOpt.equals(o.getLongOpt())) {
                return o;
            }
        }
        return null;
    }
}