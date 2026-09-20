package com.mestrap.action;

import com.mestrap.core.ActionContext;
import com.mestrap.core.JschFileUploader;
import com.mestrap.core.OverwritePolicy;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.BoolUtil;
import com.mestrap.utils.LogPrinter;
import com.mestrap.utils.VariableReplacer;
import org.apache.commons.cli.Option;
import sun.rmi.runtime.Log;

import java.io.File;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 推送本地文件至远程服务器。
 *
 * <p>-f, --file 本地文件或文件夹。
 * <p>-d, --dest 远程目标路径。
 * <p>-F, --force 远程文件已存在时是否覆盖。
 * <p>-B, --backup 覆盖前是否备份原文件。
 *
 * @author melvek
 */
public class PushAction implements TaskAction {

    @Override
    public String name() {
        return "push";
    }

    @Override
    public List<Option> cliOptions() {
        List<Option> list = new ArrayList<>();
        list.add(Option.builder("f").longOpt("file").hasArg().argName("file")
                .desc("Local file (injected as ${file})").build());
        list.add(Option.builder("d").longOpt("dest").hasArg().argName("path")
                .desc("Remote destination (injected as ${dest})").build());
        list.add(Option.builder("F").longOpt("force")
                .desc("Allow overwriting existing remote files").build());
        list.add(Option.builder("B").longOpt("backup")
                .desc("Back up existing remote file before overwrite").build());
        return list;
    }

    @Override
    public Map<String, String> cliVarMapping() {
        Map<String, String> cliVar = new HashMap<>(8);
        cliVar.put("file", "file");
        cliVar.put("dest", "dest");
        cliVar.put("force", "force");
        cliVar.put("backup", "backup");
        return cliVar;
    }

    @Override
    public void execute(ActionContext ctx) throws Exception {

        Object fileRaw = ctx.getWith().get("file");
        Object destRaw = ctx.getWith().get("dest");

        if (fileRaw == null) {
            throw new TaskException("push action requires 'file' parameter");
        }
        if (destRaw == null) {
            throw new TaskException("push action requires 'dest' parameter");
        }

        String file = VariableReplacer.replace(String.valueOf(fileRaw), ctx.getVars());
        String dest = VariableReplacer.replace(String.valueOf(destRaw), ctx.getVars());

        boolean force  = BoolUtil.isTruthy(ctx.getWith().get("force"));
        boolean backup = BoolUtil.isTruthy(ctx.getWith().get("backup"));

        File sourceFile = new File(file);
        if (!sourceFile.exists()) {
            throw new TaskException("Source file does not exist: " + sourceFile.getAbsolutePath());
        }

        OverwritePolicy policy = OverwritePolicy.of(force, backup);

        LogPrinter.info("Push file " + sourceFile.getAbsolutePath() + " to " + dest);
        LogPrinter.emptyLine();

        int exitCode = JschFileUploader.uploadFile(
                ctx.getHostVars().getHost(),
                ctx.getHostVars().getPort(),
                ctx.getHostVars().getUserName(),
                ctx.getHostVars().getPassword(),
                sourceFile.getAbsolutePath(),
                dest,
                policy
        );

        if (exitCode != 0) {
            throw new TaskException("Push failed with exit code " + exitCode);
        }
    }
}