package com.mestrap.cli;

import com.mestrap.action.TaskAction;
import com.mestrap.core.ActionRegistry;
import com.mestrap.utils.Constant;
import com.mestrap.utils.LogPrinter;
import org.apache.commons.cli.HelpFormatter;
import org.apache.commons.cli.Option;
import org.apache.commons.cli.Options;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

/**
 * 帮助信息输出。
 *
 * @author melvek
 */
public final class ShowHelp {

    private ShowHelp() {}

    /**
     * 打印版本信息。
     */
    public static void printVersion() {
        LogPrinter.section("TaskSSH v" + Constant.VERSION);
        LogPrinter.info("Lightweight SSH operations tool based on JSch");
        String buildTime = LocalDateTime.now()
                .format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
        LogPrinter.keyValue("Build time", buildTime);
    }

    /**
     * 打印全局帮助：列出所有可用 action 及其参数，以及全局选项。
     *
     * @param actionRegistry action 注册表
     */
    public static void printGlobal(ActionRegistry actionRegistry) {

        System.out.println("\nTaskSSH - Lightweight SSH operations tool");
        System.out.println("Version: " + Constant.VERSION + "\n");

        System.out.println("Usage:");
        System.out.println("  taskssh <task> [hosts...] [options]\n");

        System.out.println("Tool commands:");
        System.out.println("  encrypt    Encrypt a string");
        System.out.println("  decrypt    Decrypt a string");
        System.out.println();

        System.out.println("Available actions for task definitions:\n");

        HelpFormatter hf = new HelpFormatter();
        hf.setOptionComparator(null);
        hf.setSyntaxPrefix("Action: ");
        hf.setLeftPadding(2);
        hf.setLongOptPrefix(" --");

        for (TaskAction action : actionRegistry.all()) {
            Options opts = new Options();
            for (Option o : action.cliOptions()) {
                opts.addOption(o);
            }

            if (opts.getOptions().isEmpty()) {
                System.out.println("      (no parameters)\n");
            } else {
                hf.printHelp(action.name(), opts);
            }
            System.out.println();
        }

        System.out.println("Global options:");
        System.out.println("  -i, --inventory <file>   Inventory file (default: inventory.yaml)");
        System.out.println("  -l, --list               List target hosts only");
        System.out.println("  -P, --port     <int>     Override port");
        System.out.println("  -u, --user     <string>  Override username");
        System.out.println("  -p, --password <string>  Override password (not recommended)");
        System.out.println("  -y, --yes                Skip confirmation");
        System.out.println("  -h, --help               Show help");
        System.out.println("  -v, --version            Show version");
        System.out.println();
    }
}