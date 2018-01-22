package models

const StartupScriptHeader = `
function with_backoff {
  local max_attempts=${ATTEMPTS-5}
  local interval=${INITIAL_INTERVAL-1}
  local attempt=0
  local exitCode=0

  while (( $attempt < $max_attempts ))
  do
    set +e
    "$@"
    exitCode=$?
    set -e

    if [[ $exitCode == 0 ]]
    then
      break
    fi

    echo "Failure! Retrying in $interval.." 1>&2
    sleep $interval
    attempt=$(( attempt + 1 ))
    interval=$(( interval * 2 ))
  done

  if [[ $exitCode != 0 ]]
  then
    echo "You've failed me for the last time! ($@)" 1>&2
  fi

  return $exitCode
}
`
