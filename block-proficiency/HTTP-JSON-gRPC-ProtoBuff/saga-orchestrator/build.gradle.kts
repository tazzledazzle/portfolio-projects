//import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

plugins {
//    kotlin("jvm") version "1.9.23" apply false
    id("com.google.protobuf") version "0.9.4" apply false
}

val grpcVersion        = "1.63.0"
val grpcKotlinVersion  = "1.4.1"
val protobufVersion    = "3.25.3"
val coroutinesVersion  = "1.8.0"
val ktorVersion        = "2.3.10"

subprojects {
    apply(plugin = "org.jetbrains.kotlin.jvm")

    repositories {
        mavenCentral()
    }

    dependencies {
//        val implementation by configurations
//        val testImplementation by configurations
//
//        implementation(kotlin("stdlib"))
//        implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:$coroutinesVersion")
//        implementation("io.grpc:grpc-kotlin-stub:$grpcKotlinVersion")
//        implementation("io.grpc:grpc-protobuf:$grpcVersion")
//        implementation("io.grpc:grpc-netty-shaded:$grpcVersion")
//        implementation("com.google.protobuf:protobuf-kotlin:$protobufVersion")
//        implementation("io.github.microutils:kotlin-logging-jvm:3.0.5")
//        implementation("ch.qos.logback:logback-classic:1.5.6")
//
//        testImplementation(kotlin("test"))
//        testImplementation("io.grpc:grpc-testing:$grpcVersion")
//        testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:$coroutinesVersion")
    }


    tasks.withType<Test> { useJUnitPlatform() }
}

//tasks.withType<Copy>().configureEach {
//    duplicateStrategy = DuplicatesStrategy.WARN
//}