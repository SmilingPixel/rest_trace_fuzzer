# Check if the correct number of arguments are provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <EXECUTABLE_PATH>"
    exit 1
fi

# Set the binary path from command line arguments
EXECUTABLE_PATH="$1"

# If the binary exists, remove it
if [ -f "$EXECUTABLE_PATH" ]; then
    echo "Removing binary..."
    rm "$EXECUTABLE_PATH"
    echo "Binary removed successfully!"
else
    echo "Binary not found. Nothing to clean."
fi
