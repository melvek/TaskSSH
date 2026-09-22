package com.mestrap.cli;

import com.mestrap.exception.TaskException;
import com.mestrap.utils.EncryptTool;
import com.mestrap.utils.LogPrinter;

import java.io.Console;
import java.util.Arrays;

/**
 * 加解密子命令。
 *
 * <p>用法：
 * <pre>
 *   taskssh encrypt              # 交互输入，推荐
 *   taskssh encrypt "123456"     # 命令行传参，有泄露风险
 *   taskssh decrypt              # 交互输入
 *   taskssh decrypt "密文"
 * </pre>
 *
 * @author melvek
 */
public final class EncryptCommand {

    private EncryptCommand() {}

    /**
     * 执行加解密命令。
     *
     * @param mode 模式：encrypt 或 decrypt
     * @param args 参数（不含模式本身）
     * @return 退出码，0 成功，1 失败
     */
    public static int run(String mode, String[] args) {

        if (args.length > 1) {
            LogPrinter.error("Too many arguments");
            printUsage(mode);
            return 1;
        }

        boolean encrypt = CommandDispatcher.CMD_ENCRYPT.equals(mode);
        String input;

        if (args.length == 0) {
            // 交互输入
            input = readInteractive(encrypt);
            if (input == null) {
                return 1;
            }
        } else {
            // 命令行传参，提示风险
            input = args[0];
            LogPrinter.warning("Passing value via command line is a security risk, Use interactive input instead: taskssh " + mode);
            LogPrinter.info("Use interactive input instead: taskssh " + mode);
        }

        try {
            if (encrypt) {
                System.out.println(EncryptTool.encrypt(input));
            } else {
                System.out.println(EncryptTool.decrypt(input));
            }
            return 0;
        } catch (TaskException e) {
            LogPrinter.error(e.getMessage());
            return 1;
        }
    }

    /**
     * 从终端交互读取输入。
     *
     * @param encrypt true 表示加密，需要确认；false 表示解密
     * @return 输入内容，失败返回 null
     */
    private static String readInteractive(boolean encrypt) {
        Console console = System.console();
        if (console == null) {
            LogPrinter.error("No console available");
            LogPrinter.info("Use: taskssh " + (encrypt ? "encrypt" : "decrypt") + " \"<text>\"");
            return null;
        }

        if (encrypt) {
            char[] pwd1 = console.readPassword("Enter password: ");
            if (pwd1 == null || pwd1.length == 0) {
                LogPrinter.error("Input is empty");
                return null;
            }
            return new String(pwd1);
        }

        // 解密：输入密文
        char[] cipher = console.readPassword("Enter cipher: ");
        if (cipher == null || cipher.length == 0) {
            LogPrinter.error("Input is empty");
            return null;
        }
        String result = new String(cipher);
        Arrays.fill(cipher, ' ');
        return result;
    }

    private static void printUsage(String mode) {
        LogPrinter.info("Usage:");
        LogPrinter.info("  taskssh " + mode + "                  # interactive input");
        LogPrinter.info("  taskssh " + mode + " \"<text>\"          # command line");
    }
}