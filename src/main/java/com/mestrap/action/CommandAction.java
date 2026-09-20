package com.mestrap.action;

import com.mestrap.core.ActionContext;
import com.mestrap.ssh.JschCommandExecutor;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.LogPrinter;
import com.mestrap.utils.VariableReplacer;
import org.apache.commons.cli.Option;

import java.util.Collections;
import java.util.List;
import java.util.Map;

/**
 * 执行远程命令。
 * -e, --execute 远程执行的命令内容，支持 ${...} 格式参数
 *
 * @author melvek
 */
public class CommandAction implements TaskAction {

    @Override
    public String name() {
        return "command";
    }

    @Override
    public List<Option> cliOptions() {
        return Collections.singletonList(
                Option.builder("e").longOpt("execute").hasArg().argName("string")
                        .desc("Command string (injected as ${command})")
                        .build()
        );
    }

    @Override
    public Map<String, String> cliVarMapping() {
        return Collections.singletonMap("execute", "command");
    }

    @Override
    public void execute(ActionContext ctx) throws Exception {
        Object raw = ctx.getWith().get("command");
        if (raw == null) {
            throw new TaskException("command action requires 'command' parameter");
        }

        String cmd = VariableReplacer.replace(String.valueOf(raw), ctx.getVars());

        LogPrinter.info("Execute command: " + cmd);
        LogPrinter.emptyLine();

        int exitCode = JschCommandExecutor.executeCommand(
                ctx.getHostVars().getHost(),
                ctx.getHostVars().getPort(),
                ctx.getHostVars().getUserName(),
                ctx.getHostVars().getPassword(),
                cmd
        );

        if (exitCode != 0) {
            throw new TaskException("Command failed with exit code " + exitCode);
        }
    }
}