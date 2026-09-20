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

import java.util.Arrays;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

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

        if (args == null || args.length == 0) {
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }

        // 工具子命令：encrypt / decrypt
        if (isToolCommand(args[0])) {
            String[] remaining = Arrays.copyOfRange(args, 1, args.length);
            return EncryptCommand.run(args[0], remaining);
        }

        // 解析参数
        CommandLine cl;
        try {
            cl = parseArgs(args);
        } catch (ParseException e) {
            LogPrinter.error("Invalid arguments: " + e.getMessage());
            LogPrinter.hint("Use -h to see help");
            return 1;
        }

        // 元信息
        int meta = handleMeta(cl);
        if (meta >= 0) {
            return meta;
        }

        // 加载清单并解析任务
        Inventory inventory = loadInventory(cl);
        taskRegistry.loadTasks(inventory.getTasks());

        List<String> positional = cl.getArgList();
        if (positional.isEmpty()) {
            LogPrinter.error("No task specified");
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }
        String taskName = positional.get(0);
        Task task = taskRegistry.resolve(taskName);
        if (task == null) {
            LogPrinter.error("Unknown task: <" + taskName + ">");
            ShowHelp.printGlobal(actionRegistry);
            return 1;
        }

        // 解析目标主机
        List<String> hostNames = positional.subList(1, positional.size());
        Map<String, HostVars> hosts = HostResolver.resolve(hostNames, inventory, cl);
        if (hosts.isEmpty()) {
            LogPrinter.error("No target hosts specified");
            LogPrinter.hint("Usage: taskssh " + taskName + " <hosts...> [options]");
            return 1;
        }

        // 展示主机
        printHosts(hosts);
        if (cl.hasOption(GlobalOptions.LIST)) {
            return 0;
        }

        // 确认
        if (!confirm(cl)) {
            return 0;
        }

        // 执行
        return execute(task, hosts, inventory, cl);
    }

    // ------------------------------------------------------------------
    // 主流程辅助
    // ------------------------------------------------------------------

    /**
     * 判断是否为工具子命令。
     */
    private boolean isToolCommand(String cmd) {
        return CMD_ENCRYPT.equals(cmd) || CMD_DECRYPT.equals(cmd);
    }

    /**
     * 构建选项并解析参数。
     */
    private CommandLine parseArgs(String[] args) throws ParseException {
        Options options = new Options();
        for (Option o : GlobalOptions.all()) {
            options.addOption(o);
        }
        for (TaskAction action : actionRegistry.all()) {
            for (Option o : action.cliOptions()) {
                options.addOption(o);
            }
        }
        return new DefaultParser().parse(options, args);
    }

    /**
     * 处理元信息选项。
     *
     * @return 退出码，或 -1 表示继续执行
     */
    private int handleMeta(CommandLine cl) {
        if (cl.hasOption(GlobalOptions.VERSION)) {
            ShowHelp.printVersion();
            return 0;
        }
        if (cl.hasOption(GlobalOptions.HELP)) {
            ShowHelp.printGlobal(actionRegistry);
            return 0;
        }
        return -1;
    }

    /**
     * 加载清单文件。
     */
    private Inventory loadInventory(CommandLine cl) {
        String invFile = cl.hasOption(GlobalOptions.INVENTORY)
                ? cl.getOptionValue(GlobalOptions.INVENTORY)
                : Constant.DEFAULT_INVENTORY;
        return InventoryLoader.load(invFile);
    }

    /**
     * 展示目标主机列表。
     */
    private void printHosts(Map<String, HostVars> hosts) {
        LogPrinter.section("Target hosts");
        hosts.forEach((name, vars) -> LogPrinter.listItem(name, vars.getHost()));
        LogPrinter.emptyLine();
    }

    /**
     * 执行前确认。使用 -y 跳过。
     */
    private boolean confirm(CommandLine cl) {
        if (cl.hasOption(GlobalOptions.YES)) {
            return true;
        }
        //noinspection AlibabaUndefineMagicConstant
        return ConfirmUtil.confirm("Confirm to proceed");
    }

    /**
     * 提取 CLI 变量并执行任务。
     */
    private int execute(Task task, Map<String, HostVars> hosts,
                        Inventory inventory, CommandLine cl) {
        Map<String, Object> cliVars = extractCliVars(cl);
        Map<String, Object> globalVars = inventory.getGlobalVars() != null
                ? inventory.getGlobalVars().getExtraFields()
                : Collections.emptyMap();

        TaskExecutor executor = new TaskExecutor(actionRegistry);
        return executor.executeAll(task, hosts, globalVars, cliVars);
    }

    // ------------------------------------------------------------------
    // CLI 变量
    // ------------------------------------------------------------------

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
                if (opt == null || !cl.hasOption(cliOpt)) {
                    continue;
                }
                vars.put(varKey, opt.hasArg() ? cl.getOptionValue(cliOpt) : "true");
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