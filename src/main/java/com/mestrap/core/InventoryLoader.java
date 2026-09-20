package com.mestrap.core;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.mestrap.entity.HostVars;
import com.mestrap.entity.Inventory;
import com.mestrap.exception.TaskException;
import com.mestrap.utils.LogPrinter;

import java.io.File;
import java.io.FileInputStream;
import java.io.InputStream;
import java.text.SimpleDateFormat;
import java.util.Date;

/**
 * 清单文件加载器。
 *
 * @author melvek
 */
public final class InventoryLoader {

    /** 日期格式：yyyyMMdd */
    private static final String DATE_PATTERN = "yyyyMMdd";

    private InventoryLoader() {}

    /**
     * 加载清单文件。
     *
     * @param path 清单文件路径
     * @return 清单对象
     * @throws TaskException 文件不存在或解析失败时抛出
     */
    public static Inventory load(String path) {

        File file = new File(path);
        if (!file.exists()) {
            LogPrinter.error("Inventory not found: " + file.getAbsolutePath());
            throw new TaskException("Inventory file not found: " + file.getAbsolutePath());
        }

        try (InputStream in = new FileInputStream(file)) {
            ObjectMapper mapper = new ObjectMapper(new YAMLFactory());
            Inventory inv = mapper.readValue(in, Inventory.class);

            int groups = inv.getServers() != null ? inv.getServers().size() : 0;
            int tasks = inv.getTasks() != null ? inv.getTasks().size() : 0;
            LogPrinter.success("Loaded inventory: " + groups + " groups, " + tasks + " tasks");

            if (inv.getGlobalVars() == null) {
                inv.setGlobalVars(new HostVars());
            }

            // 注入日期变量
            SimpleDateFormat sdf = new SimpleDateFormat(DATE_PATTERN);
            inv.getGlobalVars().setExtraField("date", sdf.format(new Date()));

            return inv;

        } catch (TaskException e) {
            throw e;
        } catch (Exception e) {
            LogPrinter.error("Failed to parse inventory: " + e.getMessage());
            throw new TaskException("Failed to parse inventory: " + e.getMessage(), e);
        }
    }
}