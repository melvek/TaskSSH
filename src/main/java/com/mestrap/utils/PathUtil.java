package com.mestrap.utils;

import java.io.File;
import java.io.IOException;

/**
 * 路径处理工具。
 *
 * @author melvek
 */
public final class PathUtil {

    private PathUtil() {}

    /**
     * 取路径的文件名。
     */
    public static String getFileName(String path) {
        if (path == null || path.isEmpty()) {
            return path;
        }
        int slash = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
        return slash >= 0 ? path.substring(slash + 1) : path;
    }

    /**
     * 取路径的父目录。无父目录时返回 "/"。
     */
    public static String getParentDirectory(String path) {
        if (path == null || path.isEmpty()) {
            return Constant.SEPARATOR;
        }
        String normalized = path.replace('\\', Constant.SEPARATOR_CHAR);
        int lastSlash = normalized.lastIndexOf(Constant.SEPARATOR_CHAR);
        return lastSlash <= 0 ? Constant.SEPARATOR : normalized.substring(0, lastSlash);
    }

    /**
     * 判断路径是否以分隔符结尾（视为目录）。
     */
    public static boolean isDirectoryPath(String path) {
        return path != null && (path.endsWith("/") || path.endsWith("\\"));
    }

    /**
     * 取规范路径，失败时回退到绝对路径。
     */
    public static String canonical(File file) {
        try {
            return file.getCanonicalPath();
        } catch (IOException e) {
            return file.getAbsolutePath();
        }
    }
}