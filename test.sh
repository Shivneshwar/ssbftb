#!/bin/bash

# Define the maximum number
max_number=10  # Change this to your desired maximum number

cd ../output/

# Initialize variables to store total count and sum of values
total_count=0
total_sum=0

# Count occurrences of "Delivered message" in each output file
for ((i = 0; i < max_number; i++)); do
    output_file="output_$i.txt"

    # Use grep to count occurrences and print the result
    err_count=$(grep -c "sendto" "$output_file")

    echo "Occurrences of 'sendto' in $output_file: $err_count"
    del_count=$(grep -c "Delivered message" "$output_file")
    total_count=$((total_count + del_count))
    echo "Occurrences of 'Delivered message' in $output_file: $del_count"
done

echo "Total occurrences of 'Delivered message' in all files: $total_count"

for ((i = 0; i < max_number; i++)); do
    output_file="output_$i.txt"

    while IFS= read -r line; do
        sum=$(echo "$line" | awk '{print $NF}' | bc -l)
        # Add sum to total sum
        total_sum=$(awk "BEGIN {print $total_sum + $sum}")
    done < "$output_file"
done

echo "Total sum of values associated with occurrences: $total_sum"
