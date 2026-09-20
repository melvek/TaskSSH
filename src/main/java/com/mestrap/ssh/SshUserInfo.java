package com.mestrap.ssh;

import com.jcraft.jsch.UserInfo;
import com.mestrap.utils.EncryptTool;
import com.mestrap.utils.LogPrinter;

import java.io.Console;
import java.util.Arrays;

/**
 * SSH 认证信息。
 *
 * <p>密码来源优先级：
 * <ol>
 *   <li>清单中的加密密码</li>
 *   <li>终端交互输入</li>
 * </ol>
 *
 * <p>终端输入密码的提示格式与 OpenSSH 一致：{@code user@host's password: }
 *
 * @author melvek
 */
public class SshUserInfo implements UserInfo {

    /** 终端输入的密码缓存，避免多台主机重复输入 */
    private static String cachedConsolePassword;

    /** 解密后的密码，可为 null（表示走终端输入） */
    private String password;

    private String prompt;

    /**
     * 构造认证信息。
     *
     * @param encryptedPassword 清单中的加密密码，可为 null
     * @param username          登录用户名
     * @param host              主机地址
     */
    public SshUserInfo(String encryptedPassword) {
        if (encryptedPassword == null || encryptedPassword.isEmpty()) {
            this.password = null;
            return;
        }
        try {
            this.password = EncryptTool.decrypt(encryptedPassword);
        } catch (Exception e) {
            LogPrinter.error("Failed to decrypt password: " + e.getMessage());
            this.password = null;
        }
    }

    /**
     * 获取密码。
     *
     * <p>清单有密码则直接返回；否则从终端读取。
     */
    @Override
    public String getPassword() {
        if (password != null) {
            return password;
        }
        if (cachedConsolePassword != null) {
            return cachedConsolePassword;
        }
        cachedConsolePassword = promptFromConsole();
        return cachedConsolePassword;
    }

    @Override
    public String getPassphrase() {
        return null;
    }

    @Override
    public boolean promptPassword(String message) {
        prompt = message + ":";
        return password != null || System.console() != null;
    }

    @Override
    public boolean promptPassphrase(String message) {
        return false;
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
     * 从终端读取密码。
     *
     * @return 明文密码；无终端或输入为空时返回 null
     */
    private String promptFromConsole() {
        Console console = System.console();
        if (console == null) {
            LogPrinter.error("No password configured and no console available");
            return null;
        }
        char[] chars = console.readPassword(prompt);
        if (chars == null || chars.length == 0) {
            return null;
        }
        String pwd = new String(chars);
        Arrays.fill(chars, ' ');
        return pwd;
    }

    /**
     * 清空终端输入的密码缓存。
     *
     * <p>用于需要重新输入密码的场景。
     */
    public static void clearCachedPassword() {
        cachedConsolePassword = null;
    }
}