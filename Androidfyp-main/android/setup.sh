#!/bin/bash

# Android Project Setup Script

echo "=== Android Project Setup ==="
echo ""

# Check Java
echo "Checking Java installation..."
if command -v java &> /dev/null; then
    JAVA_VERSION=$(java -version 2>&1 | head -n 1)
    echo "✓ Java found: $JAVA_VERSION"
else
    echo "✗ Java not found. Installing OpenJDK 11..."
    sudo apt update
    sudo apt install -y openjdk-11-jdk
    export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
    export PATH=$PATH:$JAVA_HOME/bin
fi

# Set JAVA_HOME if not set
if [ -z "$JAVA_HOME" ]; then
    export JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
    export PATH=$PATH:$JAVA_HOME/bin
    echo "✓ JAVA_HOME set to: $JAVA_HOME"
fi

# Check Android SDK
echo ""
echo "Checking Android SDK..."
if [ -z "$ANDROID_HOME" ]; then
    echo "⚠ ANDROID_HOME not set"
    echo "  Please install Android Studio or Android SDK command line tools"
    echo "  Then set: export ANDROID_HOME=\$HOME/Android/Sdk"
else
    echo "✓ ANDROID_HOME: $ANDROID_HOME"
fi

# Make gradlew executable
echo ""
echo "Setting up Gradle wrapper..."
chmod +x gradlew
echo "✓ gradlew is now executable"

# Try to build
echo ""
echo "Attempting to build project..."
if ./gradlew build --no-daemon 2>&1 | tee build.log; then
    echo ""
    echo "✓ Build successful!"
    echo ""
    echo "To run on device/emulator:"
    echo "  1. Connect device: adb devices"
    echo "  2. Install: ./gradlew installDebug"
    echo "  3. Run: adb shell am start -n com.example.finalyearproject/.app.auth.view.AuthActivity"
else
    echo ""
    echo "⚠ Build failed. Check build.log for details"
    echo "Common issues:"
    echo "  - Missing Android SDK"
    echo "  - JAVA_HOME not set correctly"
    echo "  - Network issues downloading dependencies"
fi
