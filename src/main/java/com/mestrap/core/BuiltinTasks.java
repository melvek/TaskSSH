package com.mestrap.core;

import com.mestrap.entity.Step;
import com.mestrap.entity.Task;

import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * 内置任务定义。
 *
 * <p>内置任务由代码提供，用户可在清单的 tasks 段中定义同名任务覆盖。
 *
 * @author melvek
 */
public final class BuiltinTasks {

    private BuiltinTasks() {}

    /**
     * 返回所有内置任务，按注册顺序。
     *
     * @return 任务映射
     */
    public static Map<String, Task> all() {
        Map<String, Task> map = new LinkedHashMap<>(8);
        map.put("command", command());
        map.put("push", push());
        map.put("fetch", fetch());
        return map;
    }

    /**
     * 内置 command 任务：执行远程命令。
     */
    private static Task command() {
        Task task = new Task();
        task.setName("command");
        task.setDescription("Execute a remote command");

        Step step = new Step();
        step.setName("exec");
        step.setAction("command");

        Map<String, Object> withMap = new HashMap<>(2);
        withMap.put("command", "${command}");
        step.setWith(withMap);

        task.setSteps(Collections.singletonList(step));
        return task;
    }

    /**
     * 内置 push 任务：上传文件到远程。
     */
    private static Task push() {
        Task task = new Task();
        task.setName("push");
        task.setDescription("Upload a file to remote server");

        Step step = new Step();
        step.setName("upload");
        step.setAction("push");

        Map<String, Object> withMap = new HashMap<>(8);
        withMap.put("file", "${file}");
        withMap.put("dest", "${dest}");
        withMap.put("force", "${force}");
        withMap.put("backup", "${backup}");
        step.setWith(withMap);

        task.setSteps(Collections.singletonList(step));
        return task;
    }

    /**
     * 内置 fetch 任务：从远程下载文件。
     */
    private static Task fetch() {
        Task task = new Task();
        task.setName("fetch");
        task.setDescription("Download a file from remote server");

        Step step = new Step();
        step.setName("download");
        step.setAction("fetch");

        Map<String, Object> withMap = new HashMap<>(8);
        withMap.put("file", "${file}");
        withMap.put("force", "${force}");
        withMap.put("backup", "${backup}");
        step.setWith(withMap);

        task.setSteps(Collections.singletonList(step));
        return task;
    }
}