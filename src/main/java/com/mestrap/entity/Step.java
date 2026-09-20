package com.mestrap.entity;

import java.util.Map;

/**
 * @author melvek
 */
public class Step {
    private String name;
    private String action;
    private Map<String, Object> with;

    private int delay = 0;

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public String getAction() { return action; }
    public void setAction(String action) { this.action = action; }

    public Map<String, Object> getWith() { return with; }
    public void setWith(Map<String, Object> with) { this.with = with; }

    public int getDelay() {
        return delay;
    }

    public void setDelay(int delay) {
        this.delay = delay;
    }
}