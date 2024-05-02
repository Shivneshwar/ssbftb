#!/bin/bash

# Define the maximum number
max_number=10  # Change this to your desired maximum number

rm -rf ../output/*
> input.txt

# Generate lines in input.txt starting from "localhost:9000"
for ((i = 0; i < max_number; i++)); do
    printf "localhost:90%02d\n" "$i" >> input.txt
done

# Display the contents of input.txt
cat input.txt

cd code/
go build -o ../../output/brb
cd ../../output/

# Start a loop to execute the commands
for ((i = 0; i < max_number; i++)); do
    output_file="output_$i.txt"
    ./brb ../brb/input.txt "$i" > "$output_file" 2>&1 &
done

sleep 300

# Kill all the background processes
pkill -P $$  # This kills child processes of the current shell

# Optionally, you can wait for all background processes to finish
wait


# Initialize variables to store total count and sum of values
total_count=0
total_sum=0

# Count occurrences of "Delivered message" in each output file
for ((i = 0; i < max_number; i++)); do
    output_file="output_$i.txt"

    # Use grep to count occurrences and print the result
    count=$(grep -c "Delivered message" "$output_file")
    total_count=$((total_count + count))

    echo "Occurrences of 'Delivered message' in $output_file: $count"
done

echo "Total occurrences of 'Delivered message' in all files: $total_count"

for ((i = 0; i < max_number; i++)); do
    output_file="output_$i.txt"

    while IFS= read -r line; do
        sum=$(echo "$line" | grep -oP 'Delivered message \K\d+\.\d+' | awk '{ sum += $1 } END { print sum }')
        # Add sum to total sum
        total_sum=$(awk "BEGIN {print $total_sum + $sum}")
    done < "$output_file"
done

echo "Total sum of values associated with occurrences: $total_sum"


