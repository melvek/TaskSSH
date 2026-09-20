package com.mestrap;

import com.mestrap.cli.CommandDispatcher;
import com.mestrap.utils.LogPrinter;
import junit.framework.TestCase;

/**
 * @author melvek
 * @date 2026/9/20 16:33
 * @description 加解密功能测试
 */
public class EncryptTest extends TestCase {


    public void testEncrypt() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        int result = dispatcher.dispatch(new String[]{"encrypt", "123456"});

        assertEquals(0, result); // 验证返回码
    }

    public void testDecrypt() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        int result = dispatcher.dispatch(new String[]{"decrypt", "2dO7ObeRBjqyuKkMpV6Xkg=="});

        assertEquals(0, result); // 验证返回码
    }

    public void testNoArguments() {
        LogPrinter.setColorEnabled(true);

        CommandDispatcher dispatcher = new CommandDispatcher();
        int result = dispatcher.dispatch(new String[]{"decrypt"});

        assertEquals(1, result); // 验证返回码
    }
}
