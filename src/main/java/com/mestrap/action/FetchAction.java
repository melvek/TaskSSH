package com.mestrap.action;

import com.mestrap.core.ActionContext;
import com.mestrap.entity.HostVars;
import com.mestrap.exception.TaskException;
import com.mestrap.ssh.JschFileDownloader;
import com.mestrap.utils.*;
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
        String dest;
        if (destRaw == null) {
            // 默认当前目录
            dest = "./";
        } else {
            dest = VariableReplacer.replace(String.valueOf(destRaw), ctx.getVars());
        }

        if (fileRaw == null) {
            throw new TaskException("fetch action requires 'file' parameter");
        }

        String remoteFile = VariableReplacer.replace(String.valueOf(fileRaw), ctx.getVars());

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

    /**
     * 解析本地路径。
     *
     * <ul>
     *   <li>dest 以 / 结尾：dest/主机标识/文件名</li>
     *   <li>否则：dest 视为文件路径</li>
     * </ul>
     */
    private String resolveLocalPath(String dest, String remoteFile, HostVars host) {

        String fileName = PathUtil.getFileName(remoteFile);

        String path;
        if (PathUtil.isDirectoryPath(dest)) {
            String hostId = buildHostId(host);
            path = dest + hostId + File.separator + fileName;
        } else {
            path = dest;
        }

        return PathUtil.canonical(new File(path));
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