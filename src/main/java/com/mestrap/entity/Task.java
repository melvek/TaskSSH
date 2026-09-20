package com.mestrap.entity;

import java.util.List;

/**
 * @author melvek
 */
public class Task {
    private String name;
    private String description;
    private List<Step> steps;

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public List<Step> getSteps() { return steps; }
    public void setSteps(List<Step> steps) { this.steps = steps; }
}