package com.mestrap.entity;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

/**
 * @author melvek
 */
public class Inventory {

    private Map<String, ServerGroup> servers;

    @JsonProperty("global_vars")
    private HostVars globalVars;

    private Map<String, Task> tasks;

    public Map<String, ServerGroup> getServers() {
        return servers;
    }

    public void setServers(Map<String, ServerGroup> servers) {
        this.servers = servers;
    }

    public HostVars getGlobalVars() {
        return globalVars;
    }

    public void setGlobalVars(HostVars globalVars) {
        this.globalVars = globalVars;
    }

    public Map<String, Task> getTasks() { return tasks; }
    public void setTasks(Map<String, Task> tasks) { this.tasks = tasks; }
}
