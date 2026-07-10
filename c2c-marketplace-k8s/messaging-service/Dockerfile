FROM eclipse-temurin:17-jdk AS build
WORKDIR /workspace
COPY . .
RUN ./gradlew :messaging-service:installDist --no-daemon

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /workspace/messaging-service/build/install/messaging-service /app
EXPOSE 8083
ENTRYPOINT ["/app/bin/messaging-service"]
