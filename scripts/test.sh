#!/usr/bin/env bash

make -j8
for filename in tests/*.in; do
  echo ""
  echo "" 
  echo "Testing $filename"
  ./grader engine < "$filename"
done

for filename in scripts/*.in; do
  echo ""
  echo ""
  echo "Testing $filename"
  ./grader engine < "$filename"
done
