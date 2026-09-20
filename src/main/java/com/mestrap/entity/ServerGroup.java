package com.mestrap.entity;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.mestrap.deserializer.HostMapDeserializer;

import java.util.Map;

/**
 * @author melvek
 */
public class ServerGroup {
    private String groupName;

    @JsonDeserialize(using = HostMapDeserializer.class)
    private Map<String, HostVars> hosts;

    private HostVars vars = new HostVars();

    public Map<String, HostVars> getHosts() {
        return hosts;
    }

    public String getGroupName() {
        return groupName;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public void setHosts(Map<String, HostVars> hosts) {
        this.hosts = hosts;
    }

    public HostVars getVars() {
        return vars;
    }

    public void setVars(HostVars vars) {
        this.vars = vars;
    }

}
