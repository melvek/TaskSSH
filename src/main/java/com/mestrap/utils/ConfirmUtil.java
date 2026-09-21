package com.mestrap.utils;

import java.util.Locale;
import java.util.Scanner;

/**
 * @author melvek
 */
public class ConfirmUtil {

    /**
     * Ask the user whether to continue
     *
     * @param message Prompt message
     * @return true=continue, false=exit
     */
    public static boolean confirm(String message) {
        System.out.println();
        System.out.print(message + " (y/n): ");
        System.out.flush();

        Scanner scanner = new Scanner(System.in);
        while (true) {
            String input = scanner.nextLine().trim().toLowerCase(Locale.ROOT);
            if ("y".equals(input) || "yes".equals(input)) {
                return true;
            }
            if ("n".equals(input) || "no".equals(input)) {
                return false;
            }
            System.out.print("Invalid input, please enter y or n: ");
            System.out.flush();
        }
    }
}