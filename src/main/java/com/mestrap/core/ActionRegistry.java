package com.mestrap.core;

import com.mestrap.action.CommandAction;
import com.mestrap.action.FetchAction;
import com.mestrap.action.PushAction;
import com.mestrap.action.TaskAction;

import java.util.Collection;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Action 注册表。
 *
 * <p>约定：所有 Action 的 CLI 选项必须使用全局唯一的短选项名与长选项名。
 *
 * @author melvek
 */
public final class ActionRegistry {

    private final Map<String, TaskAction> actions = new LinkedHashMap<>(16);

    public ActionRegistry() {
        register(new CommandAction());
        register(new PushAction());
        register(new FetchAction());
        // 后续扩展：register(new SleepAction()); ...
    }

    /**
     * 注册一个 Action。
     *
     * @param action Action 实例
     */
    public void register(TaskAction action) {
        actions.put(action.name(), action);
    }

    /**
     * 按名查找 Action。
     *
     * @param name Action 名
     * @return Action 实例，未找到返回 null
     */
    public TaskAction get(String name) {
        return actions.get(name);
    }

    /**
     * 返回所有已注册的 Action（只读）。
     *
     * @return Action 集合，按注册顺序
     */
    public Collection<TaskAction> all() {
        return Collections.unmodifiableCollection(actions.values());
    }
}