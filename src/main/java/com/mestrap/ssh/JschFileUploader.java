package com.mestrap.ssh;

import com.jcraft.jsch.ChannelSftp;
import com.jcraft.jsch.JSch;
import com.jcraft.jsch.JSchException;
import com.jcraft.jsch.Session;
import com.jcraft.jsch.SftpException;
import com.mestrap.core.OverwritePolicy;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.Constant;
import com.mestrap.utils.EncryptTool;
import com.mestrap.utils.LogPrinter;

import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.text.SimpleDateFormat;
import java.util.Date;

/**
 * SFTP 文件上传服务。
 *
 * @author melvek
 */
public final class JschFileUploader {

    /** SSH 连接超时，毫秒 */
    private static final int CONNECT_TIMEOUT_MS = 30000;

    /** 备份文件时间戳格式 */
    private static final String BACKUP_TIMESTAMP_PATTERN = "yyyyMMddHHmmss";

    private static final ThreadLocal<SimpleDateFormat> DATE_FORMAT =
            ThreadLocal.withInitial(() -> new SimpleDateFormat(BACKUP_TIMESTAMP_PATTERN));

    private JschFileUploader() {}

    /**
     * 上传文件到远程服务器。
     *
     * @param host         主机地址
     * @param port         端口
     * @param username     用户名
     * @param password     加密后的密码
     * @param localFile    本地文件路径
     * @param remoteTarget 远程目标路径
     * @param policy       覆盖策略
     * @return 0 表示成功
     * @throws TaskException 上传失败，或目标已存在且策略为 FAIL
     */
    public static int uploadFile(String host, int port, String username,
                                 String password, String localFile,
                                 String remoteTarget, OverwritePolicy policy)
            throws TaskException {

        Session session = null;
        ChannelSftp sftp = null;

        try {
            JSch jsch = new JSch();
            session = jsch.getSession(username, host, port);

            // 支持密码认证与 keyboard-interactive 认证
            // 非完整性 keyboard-interactive，后续根据需求支持
            session.setUserInfo(new SshUserInfo(password));

            session.setConfig("StrictHostKeyChecking", "no");
            session.connect(CONNECT_TIMEOUT_MS);

            sftp = (ChannelSftp) session.openChannel("sftp");
            sftp.connect();

            // 1. 解析最终目标路径
            String finalPath = resolveTargetPath(sftp, localFile, remoteTarget);
            String parentDir = getParentDirectory(finalPath);

            // 2. 父目录必须存在
            if (!directoryExists(sftp, parentDir)) {
                throw new TaskException("Parent directory does not exist: " + parentDir);
            }

            // 3. 目标不能是目录
            if (directoryExists(sftp, finalPath)) {
                throw new TaskException("Target exists and is a directory: " + finalPath);
            }

            // 4. 根据策略处理已存在的文件
            if (exists(sftp, finalPath)) {
                handleExisting(sftp, finalPath, policy);
            }

            // 5. 上传
            try (FileInputStream fis = new FileInputStream(localFile)) {
                sftp.put(fis, finalPath, ChannelSftp.OVERWRITE);
            }

            LogPrinter.success("Upload " + localFile + " to " + finalPath + " succeeded");
            return 0;

        } catch (TaskException e) {
            throw e;
        } catch (JSchException | SftpException | IOException e) {
            throw new TaskException("Upload failed: " + e.getMessage(), e);
        } finally {
            if (sftp != null && sftp.isConnected()) {
                sftp.disconnect();
            }
            if (session != null && session.isConnected()) {
                session.disconnect();
            }
        }
    }

    // ------------------------------------------------------------------
    // 内部方法
    // ------------------------------------------------------------------

    /**
     * 根据策略处理已存在的远程文件。
     */
    private static void handleExisting(ChannelSftp sftp, String finalPath,
                                       OverwritePolicy policy) throws SftpException {

        switch (policy) {
            case FAIL:
                throw new TaskException("Remote file already exists: " + finalPath
                        + " (use -F/--force to overwrite, add -B/--backup to keep a backup)");
            case OVERWRITE:
                LogPrinter.info("Overwriting existing file: " + finalPath);
                break;
            case BACKUP:
                String backup = backupExisting(sftp, finalPath);
                LogPrinter.info("Backed up existing file to: " + backup);
                break;
            default:
        }
    }

    /**
     * 备份已存在的文件：重命名为 base.<timestamp>.ext。
     *
     * @return 备份后的完整路径
     */
    private static String backupExisting(ChannelSftp sftp, String targetPath)
            throws SftpException {

        String fileName = new File(targetPath).getName();
        String parentDir = getParentDirectory(targetPath);

        int dot = fileName.lastIndexOf('.');
        String base = dot > 0 ? fileName.substring(0, dot) : fileName;
        String ext = dot > 0 ? fileName.substring(dot) : "";

        String timestamp = DATE_FORMAT.get().format(new Date());
        String backupName = base + "." + timestamp + ext;
        String backupPath = parentDir + Constant.SEPARATOR + backupName;

        sftp.rename(targetPath, backupPath);
        return backupPath;
    }

    /**
     * 解析目标路径，模拟 cp 行为：
     * <ul>
     *   <li>目标以 / 结尾：视为目录，拼接文件名</li>
     *   <li>目标是已存在的目录：拼接文件名</li>
     *   <li>其他：视为文件路径</li>
     * </ul>
     */
    private static String resolveTargetPath(ChannelSftp sftp, String localFile,
                                            String remoteTarget) throws SftpException {

        String fileName = new File(localFile).getName();

        if (remoteTarget.endsWith(Constant.SEPARATOR)) {
            return remoteTarget + fileName;
        }

        if (exists(sftp, remoteTarget) && isDirectory(sftp, remoteTarget)) {
            return remoteTarget + Constant.SEPARATOR + fileName;
        }

        return remoteTarget;
    }

    private static boolean exists(ChannelSftp sftp, String path) {
        try {
            sftp.stat(path);
            return true;
        } catch (SftpException e) {
            return false;
        }
    }

    private static boolean isDirectory(ChannelSftp sftp, String path) {
        try {
            return sftp.stat(path).isDir();
        } catch (SftpException e) {
            return false;
        }
    }

    private static boolean directoryExists(ChannelSftp sftp, String path) {
        return isDirectory(sftp, path);
    }

    private static String getParentDirectory(String path) {
        if (path == null || path.isEmpty()) {
            return Constant.SEPARATOR;
        }
        String normalized = path.replace('\\', Constant.SEPARATOR_CHAR);
        int lastSlash = normalized.lastIndexOf(Constant.SEPARATOR_CHAR);
        return lastSlash <= 0 ? Constant.SEPARATOR : normalized.substring(0, lastSlash);
    }
}