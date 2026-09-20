package com.mestrap.exception;

/**
 * TaskSSH 统一业务异常。
 *
 * <p>携带主机名与步骤名，便于在日志和摘要中定位失败位置。
 *
 * @author melvek
 */
public class TaskException extends RuntimeException {

    /** 失败主机，可为空 */
    private final String host;

    /** 失败步骤，可为空 */
    private final String stepName;

    /**
     * 仅带消息的构造方法。
     *
     * @param message 错误消息
     */
    public TaskException(String message) {
        super(message);
        this.host = null;
        this.stepName = null;
    }

    /**
     * 带消息和原因的构造方法。
     *
     * @param message 错误消息
     * @param cause   原始异常
     */
    public TaskException(String message, Throwable cause) {
        super(message, cause);
        this.host = null;
        this.stepName = null;
    }

    /**
     * 带主机、步骤、消息的构造方法。
     *
     * @param host     失败主机
     * @param stepName 失败步骤
     * @param message  错误消息
     */
    public TaskException(String host, String stepName, String message) {
        super(message);
        this.host = host;
        this.stepName = stepName;
    }

    /**
     * 带主机、步骤、消息和原因的构造方法。
     *
     * @param host     失败主机
     * @param stepName 失败步骤
     * @param message  错误消息
     * @param cause    原始异常
     */
    public TaskException(String host, String stepName, String message, Throwable cause) {
        super(message, cause);
        this.host = host;
        this.stepName = stepName;
    }

    public String getHost() {
        return host;
    }

    public String getStepName() {
        return stepName;
    }
}