package com.mestrap.cli;

import com.mestrap.exception.TaskException;
import com.mestrap.utils.EncryptTool;
import com.mestrap.utils.LogPrinter;

/**
 * 加解密子命令。
 *
 * <p>用法：
 * <pre>
 *   taskssh encrypt "123456"
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

        if (args.length != 1) {
            LogPrinter.error("Usage: taskssh " + mode + " \"<text>\"");
            return 1;
        }

        String input = args[0];
        if (input.isEmpty()) {
            LogPrinter.error("Input is empty");
            return 1;
        }

        try {
            if (CommandDispatcher.CMD_ENCRYPT.equals(mode)) {
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
}