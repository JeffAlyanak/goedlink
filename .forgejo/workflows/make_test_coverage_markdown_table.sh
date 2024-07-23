#!/bin/bash

input_file="coverage/report"
output_file="coverage/report.md"


echo "| Coverage | Function | File | Line |" > $output_file
echo "| --- | --- | --- | --- |" >> $output_file


while IFS= read -r line
do
  if [[ $line =~ ^forge ]]; then
    path=$(echo $line | awk '{print substr($1, 34, length(substr($1, 34)) - 1)}' | awk -F ':' '{print $1}')
    line_number=$(echo $line | awk '{print $1}' | awk -F ':' '{print $2}')
    function=$(echo $line | awk '{print $2}')
    coverage=$(echo $line | awk '{print $3}')
    
    echo "| $coverage | $function | $path | $line_number |" >> $output_file
  else
    total_coverage=$(echo $line | awk '{print $3, $4}')
    echo "" >> $output_file
    echo "## Total Coverage: $total_coverage" >> $output_file
  fi
done < $input_file