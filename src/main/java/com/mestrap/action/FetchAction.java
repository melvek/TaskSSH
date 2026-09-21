package com.mestrap.action;

import com.mestrap.core.ActionContext;
import com.mestrap.entity.HostVars;
import com.mestrap.exception.TaskException;
import com.mestrap.ssh.JschFileDownloader;
import com.mestrap.utils.BoolUtil;
import com.mestrap.utils.Constant;
import com.mestrap.utils.LogPrinter;
import com.mestrap.utils.VariableReplacer;
import org.apache.commons.cli.Option;

import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 从远程主机下载文件到本地。
 *
 * <p>-f, --file 远程文件路径。
 * <p>-d, --dest 本地目标路径。
 * <p>-F, --force 本地文件已存在时是否覆盖。
 * <p>-B, --backup 覆盖前是否备份本地原文件。
 *
 * <p>本地路径规则：
 * <ul>
 *   <li>dest 以 / 结尾：视为目录，下载到 {@code dest/<主机标识>/<文件名>}</li>
 *   <li>dest 是文件路径：直接下载到该文件</li>
 * </ul>
 *
 * @author melvek
 */
public class FetchAction implements TaskAction {

    @Override
    public String name() {
        return "fetch";
    }

    @Override
    public List<Option> cliOptions() {
        List<Option> list = new ArrayList<>();
        list.add(Option.builder("f").longOpt("file").hasArg().argName("path")
                .desc("Remote file path (injected as ${file})").build());
        list.add(Option.builder("d").longOpt("dest").hasArg().argName("path")
                .desc("Local destination (injected as ${dest})").build());
        list.add(Option.builder("F").longOpt("force")
                .desc("Allow overwriting existing local files").build());
        list.add(Option.builder("B").longOpt("backup")
                .desc("Back up existing local file before overwrite").build());
        return list;
    }

    @Override
    public Map<String, String> cliVarMapping() {
        Map<String, String> map = new HashMap<>(8);
        map.put("file", "file");
        map.put("dest", "dest");
        map.put("force", "force");
        map.put("backup", "backup");
        return map;
    }

    @Override
    public void execute(ActionContext ctx) throws Exception {

        Object fileRaw = ctx.getWith().get("file");
        Object destRaw = ctx.getWith().get("dest");

        if (fileRaw == null) {
            throw new TaskException("fetch action requires 'file' parameter");
        }
        if (destRaw == null) {
            throw new TaskException("fetch action requires 'dest' parameter");
        }

        String remoteFile = VariableReplacer.replace(String.valueOf(fileRaw), ctx.getVars());
        String dest = VariableReplacer.replace(String.valueOf(destRaw), ctx.getVars());

        boolean force  = BoolUtil.isTruthy(ctx.getWith().get("force"));
        boolean backup = BoolUtil.isTruthy(ctx.getWith().get("backup"));

        String localFile = resolveLocalPath(dest, remoteFile, ctx.getHostVars());

        LogPrinter.info("Fetch " + remoteFile + " to " + localFile);

        JschFileDownloader.downloadFile(
                ctx.getHostVars(),
                remoteFile,
                localFile,
                force,
                backup
        );
    }

    // ------------------------------------------------------------------
    // 内部方法
    // ------------------------------------------------------------------

    /**
     * 解析本地路径。
     *
     * <ul>
     *   <li>dest 以 / 结尾：dest/主机标识/文件名</li>
     *   <li>否则：dest 视为文件路径</li>
     * </ul>
     */
    private String resolveLocalPath(String dest, String remoteFile, HostVars host) {

        String fileName = getFileName(remoteFile);

        // dest 不是目录：直接作为文件路径
        if (!isDirectoryDest(dest)) {
            return dest;
        }

        // dest 是目录：按主机标识分组
        String hostId = buildHostId(host);
        return dest + hostId + File.separator + fileName;
    }

    /**
     * 判断 dest 是否以路径分隔符结尾（视为目录）。
     */
    private boolean isDirectoryDest(String dest) {
        return dest.endsWith(Constant.SEPARATOR) || dest.endsWith("\\");
    }

    /**
     * 取远程路径的文件名（兼容 / 和 \）。
     */
    private String getFileName(String path) {
        int slash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
        return slash >= 0 ? path.substring(slash + 1) : path;
    }

    /**
     * 构造主机标识，用作本地目录名。
     * 主机标识中的 / 和 \ 替换为 _，避免路径层级问题。
     */
    private String buildHostId(HostVars host) {
        String hostName = host.getHost();
        if (hostName == null || hostName.isEmpty()) {
            return "unknown";
        }
        return hostName.replace('/', '_').replace('\\', '_');
    }
}