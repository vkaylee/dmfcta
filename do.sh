#!/usr/bin/env bash
# Take all envs
allEnvs=$(printenv)
# Declare an associative array 'asArr'
declare -A asArr
# Define a function to parse data from the environment variable
function parseData () {
    local mark="${1}"
    # Read each line of the environment variable
    while IFS= read -r line;
    do
        # Extract the data string from the line
        data=''
        data=${line#${mark}_[1-9]*=}
        # Extract the number from the line
        tempNumber=
        tempNumber=${line%=$data}
        number=${tempNumber#${mark}_}
        # Store the data in the associative array
        asArr[$number]="$data"
    # Extract the lines that match the environment variable format
    done <<< $(echo "${allEnvs}" | egrep "^${mark}_[1-9]+=")
}

# Scan the 'ACTION' environment variables and store them in the 'asActions' associative array
asArr=() # Reset asArr
declare -A asActions
parseData "ACTION"
for i in "${!asArr[@]}"
do
  asActions[$i]=${asArr[$i]}
done

# Scan the 'CRITERIA' environment variables and store them in the 'asCriterias' associative array
asArr=() # Reset asArr
declare -A asCriterias
parseData "CRITERIA"
for j in "${!asArr[@]}"
do
  asCriterias[$j]=${asArr[$j]}
done

# Iterate through the 'asActions' and 'asCriterias' associative arrays
for k in "${!asActions[@]}"
do
  actionFiles=()
  # Extract the criteria from the 'asCriterias' associative array
  IFS=':' read -ra criterias <<< "${asCriterias[$k]}"
  for criteria in "${criterias[@]}"; do
    # Extract the file type and directory from the criteria
    # Pattern: fileType,directory
    fileType=$(echo "${criteria}" | cut -d "," -f 1)
    directory=$(echo "${criteria}" | cut -d "," -f 2)
    
    # Find files in the directory that match the file type, and add them to the 'actionFiles' array
    while IFS= read -r filePath;
    do
      actionFiles+=("${filePath}")
    done <<< $(find "${directory}" -type f -name "${fileType}")
  done
  # Monitor file changes for the files in the 'actionFiles' array, and perform the 'asActions' action when changes are detected
  printf "%s\n" ${actionFiles[*]} | entr -s "${asActions[${k}]}"
done
