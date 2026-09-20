package com.mestrap;

import com.mestrap.cli.CommandDispatcher;
import com.mestrap.utils.LogPrinter;
import junit.framework.TestCase;

public class AppTest extends TestCase
{
    private String[] parseArgs(String argStr) {
        return argStr.split(" ");
    }

    public void testExecuteCommand() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        // 模拟用户输入 deploy 命令
        String[] args = parseArgs("command -i inventory.yaml web_master -e \"${service_path}deploy.sh\" -y");
        // String[] args = parseArgs("");

        int result = dispatcher.dispatch(args);

        assertEquals(0, result); // 验证返回码
        // 更深入的验证：可以通过Mock DeployCommand来验证execute是否被调用（见下方进阶）
    }

    public void testPushCommand() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        // 模拟用户输入 deploy 命令
        String[] args = parseArgs("push -i inventory.yaml web_master -f \"bin/taskssh-1.1.0.jar\" -d ${service_path}  -y");
        // String[] args = parseArgs("");

        int result = dispatcher.dispatch(args);

        assertEquals(0, result); // 验证返回码
        // 更深入的验证：可以通过Mock DeployCommand来验证execute是否被调用（见下方进阶）
    }

    public void testEmptyArgsShowsHelp() {
        CommandDispatcher dispatcher = new CommandDispatcher();
        int result = dispatcher.dispatch(new String[]{});
        assertEquals(0, result); // 返回非0表示出错
    }

    public void testDeployTask() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        // 模拟用户输入 deploy 命令
        String[] args = parseArgs("deploy -i inventory.yaml web_master -y");
        // String[] args = parseArgs("");

        int result = dispatcher.dispatch(args);

        assertEquals(0, result); // 验证返回码
    }

    public void testUnknowTask() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        // 模拟用户输入 deploy 命令
        String[] args = parseArgs("test -i inventory.yaml web_master -y");
        // String[] args = parseArgs("");

        int result = dispatcher.dispatch(args);

        assertEquals(0, result); // 验证返回码
    }

}
