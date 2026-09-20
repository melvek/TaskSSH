package com.mestrap.utils;

import com.mestrap.exception.TaskException;
import org.jasypt.util.text.BasicTextEncryptor;

/**
 * 加解密工具。
 *
 * <p>使用 Jasypt 的 BasicTextEncryptor，密钥硬编码。
 *
 * @author melvek
 */
public final class EncryptTool {

    /** 主密钥，请勿泄露 */
    private static final String SEC_KEY = "c7624159-ca0f-4078-9dc3-f4cd1da6a9f6";

    private EncryptTool() {}

    /**
     * 加密。
     *
     * @param plain 明文
     * @return 密文
     * @throws TaskException 明文为空或加密失败
     */
    public static String encrypt(String plain) {
        if (plain == null || plain.isEmpty()) {
            throw new TaskException("Plain text is empty");
        }
        try {
            BasicTextEncryptor encryptor = new BasicTextEncryptor();
            encryptor.setPassword(SEC_KEY);
            return encryptor.encrypt(plain);
        } catch (Exception e) {
            throw new TaskException("Encrypt failed: " + e.getMessage(), e);
        }
    }

    /**
     * 解密。
     *
     * @param cipher 密文
     * @return 明文
     * @throws TaskException 密文为空或解密失败
     */
    public static String decrypt(String cipher) {
        if (cipher == null || cipher.isEmpty()) {
            throw new TaskException("Cipher text is empty");
        }
        try {
            BasicTextEncryptor encryptor = new BasicTextEncryptor();
            encryptor.setPassword(SEC_KEY);
            return encryptor.decrypt(cipher);
        } catch (Exception e) {
            throw new TaskException("Decrypt failed: " + e.getMessage(), e);
        }
    }
}