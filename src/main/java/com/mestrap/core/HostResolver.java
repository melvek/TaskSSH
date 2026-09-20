package com.mestrap.core;

import com.mestrap.cli.GlobalOptions;
import com.mestrap.entity.HostVars;
import com.mestrap.entity.Inventory;
import com.mestrap.entity.ServerGroup;
import com.mestrap.utils.LogPrinter;
import org.apache.commons.cli.CommandLine;

import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * 目标主机解析：命令行参数 + 服务器组展开 + CLI 覆盖。
 *
 * @author melvek
 */
public final class HostResolver {

    private HostResolver() {}

    /**
     * 解析目标主机列表。
     *
     * @param hostNames 命令行中的主机名或服务器组名
     * @param inventory 清单
     * @param cl        命令行参数
     * @return 主机名到主机变量的映射
     */
    public static Map<String, HostVars> resolve(List<String> hostNames,
                                                Inventory inventory,
                                                CommandLine cl) {

        Map<String, HostVars> target = new LinkedHashMap<>(16);

        if (hostNames == null || hostNames.isEmpty()) {
            return target;
        }

        HostVars globalVars = inventory.getGlobalVars();

        for (String hostName : hostNames) {

            boolean resolved = false;

            if (inventory.getServers() != null
                    && inventory.getServers().containsKey(hostName)) {

                ServerGroup group = inventory.getServers().get(hostName);
                HostVars groupVars = group.getVars();
                if (groupVars != null) {
                    groupVars.merge(globalVars);
                }

                group.getHosts().forEach((name, vars) -> {
                    if (groupVars != null) {
                        vars.merge(groupVars);
                    }
                    String key = hostName + "/" + name;
                    target.put(key, vars);
                });
                resolved = true;
                LogPrinter.info("Server group: " + hostName
                        + " (" + group.getHosts().size() + ")");
            }

            if (!resolved) {
                HostVars vars = new HostVars();
                vars.setHost(hostName);
                if (globalVars != null) {
                    vars.setPort(globalVars.getPort());
                    vars.setUserName(globalVars.getUserName());
                    vars.setPassword(globalVars.getPassword());
                }
                target.put(hostName, vars);
                LogPrinter.info("Host: " + hostName);
            }
        }

        applyCliOverrides(target, cl);
        return target;
    }

    /**
     * 应用 CLI 覆盖：端口、用户名、密码。
     */
    private static void applyCliOverrides(Map<String, HostVars> target, CommandLine cl) {

        if (cl.hasOption(GlobalOptions.PORT)) {
            int port = Integer.parseInt(cl.getOptionValue(GlobalOptions.PORT));
            target.values().forEach(v -> v.setPort(port));
        }
        if (cl.hasOption(GlobalOptions.USERNAME)) {
            String user = cl.getOptionValue(GlobalOptions.USERNAME);
            target.values().forEach(v -> v.setUserName(user));
        }
        if (cl.hasOption(GlobalOptions.PASSWORD)) {
            String pwd = cl.getOptionValue(GlobalOptions.PASSWORD);
            target.values().forEach(v -> v.setPassword(pwd));
            LogPrinter.warning("Password passed via CLI is a security risk");
        }
    }
}