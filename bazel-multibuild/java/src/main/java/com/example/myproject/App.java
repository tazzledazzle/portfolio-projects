package com.example.myproject;

// This file is part of the Bazel Multibuild example project.
// It demonstrates a simple Java application that uses Guava's Ints class for integer comparison.
import com.google.common.primitives.Ints;

public class App {
    
  public static int compare(int a, int b) {
    return Ints.compare(a, b);
  }

  public static void main(String... args) throws Exception {
    App app = new App();
    System.out.println("Success: " + app.compare(2, 1));
  }

}
