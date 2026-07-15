FROM eclipse-temurin:17-jdk AS build
WORKDIR /workspace
COPY . .
RUN ./gradlew :payments-service:installDist --no-daemon

FROM eclipse-temurin:17-jre-jammy
WORKDIR /app
COPY --from=build /workspace/payments-service/build/install/payments-service /app
EXPOSE 8084
ENTRYPOINT ["/app/bin/payments-service"]
