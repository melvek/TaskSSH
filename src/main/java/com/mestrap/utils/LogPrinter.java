package com.mestrap.utils;

/**
 * Colored console output utility class
 * Supports ANSI color codes; automatically disables colors on Windows
 */
public class LogPrinter {

    /** 系统属性键：操作系统名称 */
    private static final String PROP_OS_NAME = "os.name";

    /** 操作系统名称中表示 Windows 的关键字 */
    private static final String OS_WINDOWS = "windows";

    /** 环境变量：强制启用颜色 */
    private static final String ENV_FORCE_COLOR = "TaskSSH_COLOR";

    private static final String RESET = "\033[0m";
    private static final String BLACK = "\033[0;30m";
    private static final String RED = "\033[0;31m";
    private static final String GREEN = "\033[0;32m";
    private static final String YELLOW = "\033[0;33m";
    private static final String BLUE = "\033[0;34m";
    private static final String PURPLE = "\033[0;35m";
    private static final String CYAN = "\033[0;36m";
    private static final String WHITE = "\033[0;37m";

    // Bold

    private static final String BOLD = "\033[1m";
    private static final String BOLD_RED = "\033[1;31m";
    private static final String BOLD_GREEN = "\033[1;32m";
    private static final String BOLD_YELLOW = "\033[1;33m";
    private static final String BOLD_BLUE = "\033[1;34m";

    /** Whether color is enabled (auto-detected) */
    private static boolean colorEnabled = true;

    static {
        // Detect whether color is supported
        detectColorSupport();
    }

    /**
     * Detect whether the current terminal supports color
     */
    private static void detectColorSupport() {
        // Windows does not support ANSI colors by default (unless enabled)
        String os = System.getProperty(PROP_OS_NAME).toLowerCase();
        if (os.contains(OS_WINDOWS)) {
            // Windows 10 and above may support it, but it is disabled by default
            // It can be enabled as needed; here it is disabled by default for compatibility
            colorEnabled = false;

            // If an environment variable is set, it can be forcibly enabled
            if (System.getenv(ENV_FORCE_COLOR) != null) {
                colorEnabled = true;
            }
        }

        // If in a non-interactive terminal (such as CI/CD), disable color
        if (System.console() == null) {
            colorEnabled = false;
        }
    }

    /**
     * Enable or disable color output
     */
    public static void setColorEnabled(boolean enabled) {
        colorEnabled = enabled;
    }

    /**
     * Get whether color is currently enabled
     */
    public static boolean isColorEnabled() {
        return colorEnabled;
    }

    // ==================== Basic output methods ====================

    /**
     * Normal information (blue)
     */
    public static void info(String message) {
        System.out.println(message);
    }

    /**
     * Success information (green)
     */
    public static void success(String message) {
        if (colorEnabled) {
            System.out.println(GREEN + message + RESET);
        } else {
            System.out.println("[SUCCESS] " + message);
        }
    }

    /**
     * Error information (red)
     */
    public static void error(String message) {
        if (colorEnabled) {
            System.err.println(RED + "[ERROR] " + message + RESET);
        } else {
            System.err.println("[ERROR] " + message);
        }
    }

    /**
     * Error information (with detailed stack trace)
     */
    public static void error(String message, Throwable e) {
        if (colorEnabled) {
            System.err.println(RED + "[ERROR] " + message + RESET);
            if (e != null) {
                System.err.println(RED + "  - " + e.getMessage() + RESET);
            }
        } else {
            System.err.println("[ERROR] " + message);
            if (e != null) {
                System.err.println("  - " + e.getMessage());
            }
        }
    }

    /**
     * Warning information (yellow)
     */
    public static void warning(String message) {
        if (colorEnabled) {
            System.out.println(YELLOW + "[WARN] " + message + RESET);
        } else {
            System.out.println("[WARN] " + message);
        }
    }

    /**
     * Warning information (with warning symbol)
     */
    public static void warningWithSymbol(String message) {
        if (colorEnabled) {
            System.out.println(YELLOW + "⚠ " + message + RESET);
        } else {
            System.out.println("[WARN] " + message);
        }
    }

    /**
     * Hint information (cyan)
     */
    public static void hint(String message) {
        if (colorEnabled) {
            System.out.println(CYAN + "  ↳ " + message + RESET);
        } else {
            System.out.println("  ↳ " + message);
        }
    }

    /**
     * Debug information (purple)
     */
    public static void debug(String message) {
        if (colorEnabled) {
            System.out.println(PURPLE + "[DEBUG] " + message + RESET);
        } else {
            System.out.println("[DEBUG] " + message);
        }
    }

    // ==================== Structured output ====================

    /**
     * Print a divider title
     */
    public static void section(String title) {
        if (colorEnabled) {
            System.out.println("\n" + BOLD_BLUE + "=== " + title + " ===" + RESET);
        } else {
            System.out.println("\n=== " + title + " ===");
        }
    }

    /**
     * Print progress (updated on the same line)
     */
    public static void progress(int current, int total, String message) {
        String text = String.format("[%d/%d] %s", current, total, message);
        if (colorEnabled) {
            System.out.println("\r" + CYAN + text + RESET);
        } else {
            System.out.println("\r" + text);
        }
        if (current == total) {
            // System.out.println(); // New line after completion
        }
    }

    /**
     * Print a list item
     */
    public static void listItem(String key, String value) {
        if (colorEnabled) {
            System.out.printf("  %-20s -> %s%n",
                    BOLD + key + RESET,
                    CYAN + value + RESET);
        } else {
            System.out.printf("  %-20s -> %s%n", key, value);
        }
    }

    /**
     * Print a key-value pair
     */
    public static void keyValue(String key, Object value) {
        if (colorEnabled) {
            System.out.printf("  %s: %s%n",
                    BOLD + key + RESET,
                    CYAN + value + RESET);
        } else {
            System.out.printf("  %s: %s%n", key, value);
        }
    }

    /**
     * Print execution summary
     */
    public static void summary(int total, int success, int failed, String... failedHosts) {
        section("Execution Summary");
        keyValue("Total tasks", total);
        keyValue("Succeeded", success);

        if (failed > 0) {
            if (colorEnabled) {
                System.out.printf("  %s: %d %n", BOLD_RED + "Failed" + RESET, failed);
            } else {
                System.out.printf("  Failed: %d%n", failed);
            }
            if (failedHosts != null && failedHosts.length > 0) {
                hint("Failed hosts: " + String.join(", ", failedHosts));
            }
        } else {
            keyValue("Failed", "0");
        }

        // Timestamp
        String time = java.time.LocalDateTime.now()
                .format(java.time.format.DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"));
        keyValue("Completion time", time);
        System.out.println();
    }

    // ==================== Advanced formatting ====================

    /**
     * Print a colored message (custom color)
     */
    public static void print(String message, Color color) {
        if (colorEnabled && color != null) {
            System.out.println(color.code + message + RESET);
        } else {
            System.out.println(message);
        }
    }

    /**
     * Print a colored message (custom color, no newline)
     */
    public static void printNoNewline(String message, Color color) {
        if (colorEnabled && color != null) {
            System.out.print(color.code + message + RESET);
        } else {
            System.out.print(message);
        }
    }

    /**
     * Print table header
     */
    public static void tableHeader(String... headers) {
        if (colorEnabled) {
            System.out.print(BOLD);
            for (String header : headers) {
                System.out.printf("%-20s", header);
            }
            System.out.println(RESET);
        } else {
            for (String header : headers) {
                System.out.printf("%-20s", header);
            }
            System.out.println();
        }
    }

    /**
     * Print table row
     */
    public static void tableRow(String... columns) {
        for (String col : columns) {
            System.out.printf("%-20s", col);
        }
        System.out.println();
    }

    /**
     * Print horizontal line
     */
    public static void line() {
        if (colorEnabled) {
            System.out.println(CYAN + "----------------------------------------" + RESET);
        } else {
            System.out.println("----------------------------------------");
        }
    }

    /**
     * Print an empty line
     */
    public static void emptyLine() {
        System.out.println();
    }

    // ==================== Custom color enum ====================

    public enum Color {
        // 黑色
        BLACK(LogPrinter.BLACK),
        RED(LogPrinter.RED),
        GREEN(LogPrinter.GREEN),
        YELLOW(LogPrinter.YELLOW),
        BLUE(LogPrinter.BLUE),
        PURPLE(LogPrinter.PURPLE),
        CYAN(LogPrinter.CYAN),
        WHITE(LogPrinter.WHITE),
        BOLD(LogPrinter.BOLD),
        BOLD_RED(LogPrinter.BOLD_RED),
        BOLD_GREEN(LogPrinter.BOLD_GREEN),
        BOLD_YELLOW(LogPrinter.BOLD_YELLOW),
        BOLD_BLUE(LogPrinter.BOLD_BLUE);

        private final String code;

        Color(String code) {
            this.code = code;
        }

        public String getCode() {
            return code;
        }
    }
}