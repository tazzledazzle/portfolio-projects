plugins {
    id("com.tazzledazzle.python") version "0.2.0"
}


tasks.register("buildOnboardingAutomation").configure {
    dependsOn("onboarding-automation-cli")
}