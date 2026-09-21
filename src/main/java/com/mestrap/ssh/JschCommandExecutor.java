package com.mestrap.ssh;

import com.jcraft.jsch.ChannelExec;
import com.jcraft.jsch.JSchException;
import com.jcraft.jsch.Session;
import com.mestrap.entity.HostVars;
import com.mestrap.utils.LogPrinter;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;

/**
 * 远程命令执行服务。
 *
 * @author melvek
 */
public final class JschCommandExecutor {

    /** 等待 channel 关闭的轮询间隔，毫秒 */
    private static final long CHANNEL_POLL_INTERVAL_MS = 50L;

    /** 连接异常时的退出码 */
    private static final int EXIT_CODE_ERROR = -1;

    private JschCommandExecutor() {}

    /**
     * 执行远程命令。
     *
     * @param host    目标主机
     * @param command 命令内容
     * @return 退出码：0 成功，非 0 失败，-1 连接或执行异常
     */
    public static int executeCommand(HostVars host, String command) {

        Session session = null;
        ChannelExec channel = null;

        try {
            session = SshSessionFactory.connect(host);

            channel = (ChannelExec) session.openChannel("exec");
            channel.setCommand(command + " 2>&1");

            InputStream in = channel.getInputStream();
            channel.connect();

            BufferedReader reader = new BufferedReader(
                    new InputStreamReader(in, StandardCharsets.UTF_8));
            String line;
            while ((line = reader.readLine()) != null) {
                LogPrinter.info(line);
            }

            while (!channel.isClosed()) {
                Thread.sleep(CHANNEL_POLL_INTERVAL_MS);
            }

            return channel.getExitStatus();

        } catch (JSchException | IOException | InterruptedException e) {
            LogPrinter.error("Execution error: " + e.getMessage());
            if (e instanceof InterruptedException) {
                Thread.currentThread().interrupt();
            }
            return EXIT_CODE_ERROR;
        } finally {
            if (channel != null && channel.isConnected()) {
                channel.disconnect();
            }
            if (session != null && session.isConnected()) {
                session.disconnect();
            }
        }
    }
}