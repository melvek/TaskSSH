package com.mestrap.utils;

import com.mestrap.exception.TaskException;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * 变量替换工具。
 *
 * <p>严格模式：未解析的 {@code ${var}} 会抛出 {@link TaskException}。
 * <p>递归展开最多 4 层，覆盖 global / group / host / step / CLI 五个来源。
 *
 * @author melvek
 */
public final class VariableReplacer {

    private static final Pattern PLACEHOLDER = Pattern.compile("\\$\\{([^}]+)\\}");
    private static final int MAX_DEPTH = 4;

    private VariableReplacer() {}

    /**
     * 替换变量。
     *
     * @param template 模板字符串，可含 ${var}
     * @param vars     变量池
     * @return 替换后的字符串
     * @throws TaskException 存在未解析的 ${var} 时抛出
     */
    public static String replace(String template, Map<String, Object> vars) {
        if (template == null || template.isEmpty()) {
            return template;
        }

        String current = template;
        for (int i = 0; i < MAX_DEPTH; i++) {
            String next = replaceOnce(current, vars);
            if (next.equals(current)) {
                break;
            }
            current = next;
        }

        List<String> missing = findMissingKeys(current);
        if (!missing.isEmpty()) {
            throw new TaskException(
                    "Unresolved variable(s): " + missing + " in \"" + template + "\"");
        }

        return current;
    }

    /**
     * 单次替换，把 ${var} 用 vars 里的值替换掉，未命中的保留原样。
     */
    private static String replaceOnce(String template, Map<String, Object> vars) {
        Matcher m = PLACEHOLDER.matcher(template);
        StringBuffer sb = new StringBuffer();
        while (m.find()) {
            String key = m.group(1).trim();
            String value = vars.containsKey(key)
                    ? String.valueOf(vars.get(key))
                    : m.group(0);
            m.appendReplacement(sb, Matcher.quoteReplacement(value));
        }
        m.appendTail(sb);
        return sb.toString();
    }

    /**
     * 找出字符串中所有未解析的占位符键名。
     */
    private static List<String> findMissingKeys(String s) {
        List<String> keys = new ArrayList<>();
        Matcher m = PLACEHOLDER.matcher(s);
        while (m.find()) {
            keys.add(m.group(1).trim());
        }
        return keys;
    }
}