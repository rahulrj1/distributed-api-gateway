FROM maven:3.9-eclipse-temurin-17

WORKDIR /app

# Copy pom.xml and download dependencies (cached layer)
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy source
COPY src ./src

# Build the jar
RUN mvn package -DskipTests -B

# Expose port
EXPOSE 6001

# Run the application
CMD ["java", "-jar", "target/service-b-1.0.0.jar"]
