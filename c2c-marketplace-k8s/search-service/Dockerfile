FROM eclipse-temurin:17-jdk AS build
WORKDIR /workspace
COPY . .
RUN ./gradlew :search-service:installDist --no-daemon

FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY --from=build /workspace/search-service/build/install/search-service /app
EXPOSE 8082
ENTRYPOINT ["/app/bin/search-service"]
