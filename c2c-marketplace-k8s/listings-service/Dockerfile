# Multi-stage: build with Gradle, run on a slim JRE
FROM eclipse-temurin:17-jdk AS build
WORKDIR /workspace
COPY . .
RUN ./gradlew :listings-service:installDist --no-daemon

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /workspace/listings-service/build/install/listings-service /app
EXPOSE 8081
ENTRYPOINT ["/app/bin/listings-service"]
