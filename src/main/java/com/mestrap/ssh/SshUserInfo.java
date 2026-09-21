package com.mestrap.ssh;

import com.jcraft.jsch.UserInfo;
import com.mestrap.utils.EncryptTool;
import com.mestrap.utils.LogPrinter;

import java.io.Console;
import java.util.Arrays;

/**
 * SSH 认证信息。
 *
 * <p>密码与口令来源优先级：
 * <ol>
 *   <li>清单中的加密值</li>
 *   <li>终端交互输入</li>
 * </ol>
 *
 * <p>终端输入提示格式与 OpenSSH 一致：{@code user@host's password: }
 *
 * @author melvek
 */
public class SshUserInfo implements UserInfo {

    /** 终端输入的密码缓存，避免多台主机重复输入 */
    private static String cachedConsolePassword;

    /** 终端输入的口令缓存 */
    private static String cachedConsolePassphrase;

    /** 解密后的密码，可为 null（表示走终端输入） */
    private final String password;

    /** 解密后的口令，可为 null */
    private final String passphrase;

    /** 终端提示语，格式：user@host's password: */
    private final String passwordPrompt;

    /**
     * 构造认证信息。
     *
     * @param encryptedPassword   清单中的加密密码，可为 null
     * @param encryptedPassphrase 清单中的加密口令，可为 null
     * @param username            登录用户名
     * @param host                主机地址
     */
    public SshUserInfo(String encryptedPassword, String encryptedPassphrase,
                       String username, String host) {
        this.password = decryptOrNull(encryptedPassword);
        this.passphrase = decryptOrNull(encryptedPassphrase);
        this.passwordPrompt = username + "@" + host + "'s password: ";
    }

    @Override
    public String getPassword() {
        if (password != null) {
            return password;
        }
        if (cachedConsolePassword != null) {
            return cachedConsolePassword;
        }
        cachedConsolePassword = promptFromConsole(passwordPrompt);
        return cachedConsolePassword;
    }

    @Override
    public String getPassphrase() {
        if (passphrase != null) {
            return passphrase;
        }
        if (cachedConsolePassphrase != null) {
            return cachedConsolePassphrase;
        }
        cachedConsolePassphrase = promptFromConsole("Passphrase: ");
        return cachedConsolePassphrase;
    }

    @Override
    public boolean promptPassword(String message) {
        return password != null || System.console() != null;
    }

    @Override
    public boolean promptPassphrase(String message) {
        return passphrase != null || System.console() != null;
    }

    @Override
    public boolean promptYesNo(String message) {
        return true;
    }

    @Override
    public void showMessage(String message) {
        // 忽略
    }

    /**
     * 清空终端输入的密码与口令缓存。
     */
    public static void clearCache() {
        cachedConsolePassword = null;
        cachedConsolePassphrase = null;
    }

    // ------------------------------------------------------------------
    // 内部方法
    // ------------------------------------------------------------------

    private static String decryptOrNull(String encrypted) {
        if (encrypted == null || encrypted.isEmpty()) {
            return null;
        }
        try {
            return EncryptTool.decrypt(encrypted);
        } catch (Exception e) {
            LogPrinter.error("Failed to decrypt: " + e.getMessage());
            return null;
        }
    }

    private static String promptFromConsole(String prompt) {
        Console console = System.console();
        if (console == null) {
            LogPrinter.error("No console available");
            return null;
        }
        char[] chars = console.readPassword(prompt);
        if (chars == null || chars.length == 0) {
            return null;
        }
        String value = new String(chars);
        Arrays.fill(chars, ' ');
        return value;
    }
}