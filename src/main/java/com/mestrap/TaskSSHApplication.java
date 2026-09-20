package com.mestrap;

import com.mestrap.cli.CommandDispatcher;

/**
 * TaskSSH is a lightweight operations tool developed based on JSch, used for batch uploading files to remote servers and executing remote commands.
 * TaskSSHApplication 里 SSH 是全大写，Alibaba 认为应该写 TaskSshApplication，鉴于项目可读性，维持原样，特加此抑制注释
 * @author melvek
 */
@SuppressWarnings("AlibabaClassNamingShouldBeCamel")
public class TaskSSHApplication {
    public static void main( String[] args ) {
        CommandDispatcher dispatcher = new CommandDispatcher();
        int exitCode = dispatcher.dispatch(args);
        System.exit(exitCode);
    }
}