package com.mestrap.ssh;

import com.jcraft.jsch.ChannelSftp;
import com.jcraft.jsch.JSchException;
import com.jcraft.jsch.Session;
import com.jcraft.jsch.SftpException;
import com.mestrap.entity.HostVars;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.LogPrinter;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.text.SimpleDateFormat;
import java.util.Date;

/**
 * SFTP 文件下载服务。
 *
 * @author melvek
 */
public final class JschFileDownloader {

    private static final String BACKUP_TIMESTAMP_PATTERN = "yyyyMMddHHmmss";

    private static final ThreadLocal<SimpleDateFormat> DATE_FORMAT =
            ThreadLocal.withInitial(() -> new SimpleDateFormat(BACKUP_TIMESTAMP_PATTERN));

    private JschFileDownloader() {}

    /**
     * 从远程下载文件到本地。
     */
    public static int downloadFile(HostVars host, String remoteFile,
                                   String localFile, boolean force, boolean backup)
            throws TaskException {

        Session session = null;
        ChannelSftp sftp = null;

        try {
            session = SshSessionFactory.connect(host);
            sftp = (ChannelSftp) session.openChannel("sftp");
            sftp.connect();

            if (!exists(sftp, remoteFile)) {
                throw new TaskException("Remote file does not exist: " + remoteFile);
            }
            if (isDirectory(sftp, remoteFile)) {
                throw new TaskException("Remote path is a directory, not a file: "
                        + remoteFile + " (please pack it first)");
            }

            File local = new File(localFile);
            ensureParentDir(local);

            if (local.exists()) {
                if (!force) {
                    throw new TaskException("Local file already exists: "
                            + local.getAbsolutePath()
                            + " (use -F/--force to overwrite, add -B/--backup to keep a backup)");
                }
                if (backup) {
                    String backupPath = backupLocal(local);
                    LogPrinter.info("Backed up local file to: " + backupPath);
                } else {
                    LogPrinter.info("Overwriting local file: " + local.getAbsolutePath());
                }
            }

            try (InputStream in = sftp.get(remoteFile);
                 FileOutputStream out = new FileOutputStream(local)) {
                byte[] buf = new byte[8192];
                int n;
                while ((n = in.read(buf)) != -1) {
                    out.write(buf, 0, n);
                }
            }

            LogPrinter.success("Downloaded: " + remoteFile + " -> " + local.getAbsolutePath());
            return 0;

        } catch (TaskException e) {
            throw e;
        } catch (JSchException | SftpException | IOException e) {
            throw new TaskException("Download failed: " + e.getMessage(), e);
        } finally {
            if (sftp != null && sftp.isConnected()) {
                sftp.disconnect();
            }
            if (session != null && session.isConnected()) {
                session.disconnect();
            }
        }
    }

    private static void ensureParentDir(File file) throws TaskException {
        File parent = file.getParentFile();
        if (parent != null && !parent.exists()) {
            if (!parent.mkdirs()) {
                throw new TaskException("Failed to create local directory: " + parent);
            }
        }
    }

    private static String backupLocal(File file) throws TaskException {
        String name = file.getName();
        int dot = name.lastIndexOf('.');
        String base = dot > 0 ? name.substring(0, dot) : name;
        String ext = dot > 0 ? name.substring(dot) : "";

        String timestamp = DATE_FORMAT.get().format(new Date());
        String backupName = base + "." + timestamp + ext;
        File backup = new File(file.getParentFile(), backupName);

        if (!file.renameTo(backup)) {
            throw new TaskException("Failed to backup local file: " + file);
        }
        return backup.getAbsolutePath();
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
}