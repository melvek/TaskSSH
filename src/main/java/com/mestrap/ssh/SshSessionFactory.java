package com.mestrap.ssh;

import com.jcraft.jsch.JSch;
import com.jcraft.jsch.JSchException;
import com.jcraft.jsch.Session;
import com.mestrap.entity.HostVars;
import com.mestrap.utils.LogPrinter;

/**
 * SSH 会话工厂。
 *
 * <p>根据 HostVars 创建并连接 Session，统一处理：
 * <ul>
 *   <li>认证方式：密码 / 私钥 / 私钥口令 / 终端交互</li>
 *   <li>主机密钥检查：跳过（StrictHostKeyChecking=no）</li>
 *   <li>连接超时</li>
 * </ul>
 *
 * <p>支持的 extraFields 字段：
 * <ul>
 *   <li>{@code identity_file}：私钥文件路径，支持 ~ 展开</li>
 *   <li>{@code passphrase}：私钥口令（加密）</li>
 * </ul>
 *
 * @author melvek
 */
public final class SshSessionFactory {

    /** SSH 连接超时，毫秒 */
    private static final int CONNECT_TIMEOUT_MS = 30000;

    /** extraFields 中的私钥路径键 */
    private static final String EXTRA_IDENTITY_FILE = "identity_file";

    /** extraFields 中的口令键 */
    private static final String EXTRA_PASSPHRASE = "passphrase";

    /** 认证顺序 */
    private static final String PREFERRED_AUTH = "publickey,keyboard-interactive,password";

    private SshSessionFactory() {}

    /**
     * 创建并连接 Session。
     *
     * @param host 主机信息
     * @return 已连接的 Session，调用方负责 disconnect
     * @throws JSchException 连接失败
     */
    public static Session connect(HostVars host) throws JSchException {

        JSch jsch = new JSch();

        // 1. 加载私钥（如果配置）
        String identityFile = (String) host.getExtra(EXTRA_IDENTITY_FILE);
        if (identityFile != null && !identityFile.isEmpty()) {
            String expanded = expandHome(identityFile);
            // 不传 passphrase，由 SshUserInfo.getPassphrase() 提供
            jsch.addIdentity(expanded);
            LogPrinter.debug("Loaded identity: " + expanded);
        }

        // 2. 创建 Session
        if (host.getPort() == null) {
            host.setPort(22);
        }

        Session session = jsch.getSession(host.getUserName(), host.getHost(), host.getPort());

        // 3. 设置 UserInfo（密码 + 口令）
        String encryptedPassphrase = (String) host.getExtra(EXTRA_PASSPHRASE);
        session.setUserInfo(new SshUserInfo(
                host.getPassword(),
                encryptedPassphrase,
                host.getUserName(),
                host.getHost()));

        // 4. 跳过主机密钥检查
        session.setConfig("StrictHostKeyChecking", "no");

        // 5. 指定认证顺序
        session.setConfig("PreferredAuthentications", PREFERRED_AUTH);

        // 6. 连接
        session.connect(CONNECT_TIMEOUT_MS);
        return session;
    }

    /**
     * 展开 ~ 为当前用户 home 目录。
     *
     * @param path 原始路径
     * @return 展开后的路径
     */
    private static String expandHome(String path) {
        //noinspection AlibabaUndefineMagicConstant
        if (path.startsWith("~/")) {
            return System.getProperty("user.home") + path.substring(1);
        }
        return path;
    }
}