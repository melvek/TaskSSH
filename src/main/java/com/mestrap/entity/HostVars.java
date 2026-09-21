package com.mestrap.entity;

import com.fasterxml.jackson.annotation.JsonAnyGetter;
import com.fasterxml.jackson.annotation.JsonAnySetter;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.HashMap;
import java.util.Map;

/**
 * 服务器主机信息。
 *
 * <p>包含主机地址、端口、认证信息，以及任意扩展字段。
 * 除 host / port / username / password 外的字段进入 extraFields，
 * 可用于变量替换。
 *
 * @author melvek
 */
public class HostVars {

    private String host;

    private Integer port;

    @JsonProperty("username")
    private String userName;

    private String password;

    /** 扩展字段，存放自定义变量 */
    private Map<String, Object> extraFields = new HashMap<>(16);

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public Integer getPort() {
        return port;
    }

    public void setPort(Integer port) {
        this.port = port;
    }

    public String getUserName() {
        return userName;
    }

    public void setUserName(String userName) {
        this.userName = userName;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }

    @JsonAnyGetter
    public Map<String, Object> getExtraFields() {
        return extraFields;
    }

    public void setExtraFields(Map<String, Object> extraFields) {
        this.extraFields = extraFields;
    }

    /**
     * 捕获未定义的字段，存入 extraFields。
     */
    @JsonAnySetter
    public void setExtraField(String key, Object value) {
        extraFields.put(key, value);
    }

    /**
     * 获取扩展字段的值。
     *
     * @param key 字段名
     * @return 字段值，不存在返回 null
     */
    public Object getExtra(String key) {
        return extraFields.get(key);
    }

    /**
     * 合并另一个 HostVars。
     *
     * <p>规则：目标已有值则跳过，不覆盖。
     * 合并范围：port / userName / password / extraFields。
     *
     * @param source 源对象
     */
    public void merge(HostVars source) {
        if (source == null) {
            return;
        }

        if (source.getPort() != null && this.port == null) {
            this.port = source.getPort();
        }
        if (isNotEmpty(source.getUserName())
                && !isNotEmpty(this.userName)) {
            this.userName = source.getUserName();
        }
        if (isNotEmpty(source.getPassword())
                && !isNotEmpty(this.password)) {
            this.password = source.getPassword();
        }

        Map<String, Object> sourceExtra = source.getExtraFields();
        if (sourceExtra != null && !sourceExtra.isEmpty()) {
            for (Map.Entry<String, Object> entry : sourceExtra.entrySet()) {
                if (!this.extraFields.containsKey(entry.getKey())) {
                    this.extraFields.put(entry.getKey(), entry.getValue());
                }
            }
        }
    }

    private static boolean isNotEmpty(String s) {
        return s != null && !s.isEmpty();
    }
}