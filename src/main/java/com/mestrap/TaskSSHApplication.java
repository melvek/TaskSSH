package com.mestrap;

import com.mestrap.cli.CommandDispatcher;

/**
 * TaskSSH is a lightweight operations tool developed based on JSch, used for batch uploading files to remote servers and executing remote commands.
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