# bash completion for kafkactl                             -*- shell-script -*-

__kafkactl_debug()
{
    if [[ -n ${BASH_COMP_DEBUG_FILE} ]]; then
        echo "$*" >> "${BASH_COMP_DEBUG_FILE}"
    fi
}

# Homebrew on Macs have version 1.3 of bash-completion which doesn't include
# _init_completion. This is a very minimal version of that function.
__kafkactl_init_completion()
{
    COMPREPLY=()
    _get_comp_words_by_ref "$@" cur prev words cword
}

__kafkactl_index_of_word()
{
    local w word=$1
    shift
    index=0
    for w in "$@"; do
        [[ $w = "$word" ]] && return
        index=$((index+1))
    done
    index=-1
}

__kafkactl_contains_word()
{
    local w word=$1; shift
    for w in "$@"; do
        [[ $w = "$word" ]] && return
    done
    return 1
}

__kafkactl_handle_go_custom_completion()
{
    __kafkactl_debug "${FUNCNAME[0]}: cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}"

    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16

    local out requestComp lastParam lastChar comp directive args

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly kafkactl allows to handle aliases
    args=("${words[@]:1}")
    requestComp="${words[0]} __completeNoDesc ${args[*]}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __kafkactl_debug "${FUNCNAME[0]}: lastParam ${lastParam}, lastChar ${lastChar}"

    if [ -z "${cur}" ] && [ "${lastChar}" != "=" ]; then
        # If the last parameter is complete (there is a space following it)
        # We add an extra empty parameter so we can indicate this to the go method.
        __kafkactl_debug "${FUNCNAME[0]}: Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __kafkactl_debug "${FUNCNAME[0]}: calling ${requestComp}"
    # Use eval to handle any environment variables and such
    out=$(eval "${requestComp}" 2>/dev/null)

    # Extract the directive integer at the very end of the output following a colon (:)
    directive=${out##*:}
    # Remove the directive
    out=${out%:*}
    if [ "${directive}" = "${out}" ]; then
        # There is not directive specified
        directive=0
    fi
    __kafkactl_debug "${FUNCNAME[0]}: the completion directive is: ${directive}"
    __kafkactl_debug "${FUNCNAME[0]}: the completions are: ${out[*]}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        # Error code.  No completion.
        __kafkactl_debug "${FUNCNAME[0]}: received error from custom completion go code"
        return
    else
        if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kafkactl_debug "${FUNCNAME[0]}: activating no space"
                compopt -o nospace
            fi
        fi
        if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
            if [[ $(type -t compopt) = "builtin" ]]; then
                __kafkactl_debug "${FUNCNAME[0]}: activating no file completion"
                compopt +o default
            fi
        fi
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        # File extension filtering
        local fullFilter filter filteringCmd
        # Do not use quotes around the $out variable or else newline
        # characters will be kept.
        for filter in ${out[*]}; do
            fullFilter+="$filter|"
        done

        filteringCmd="_filedir $fullFilter"
        __kafkactl_debug "File filtering command: $filteringCmd"
        $filteringCmd
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        # File completion for directories only
        local subDir
        # Use printf to strip any trailing newline
        subdir=$(printf "%s" "${out[0]}")
        if [ -n "$subdir" ]; then
            __kafkactl_debug "Listing directories in $subdir"
            __kafkactl_handle_subdirs_in_dir_flag "$subdir"
        else
            __kafkactl_debug "Listing directories in ."
            _filedir -d
        fi
    else
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${out[*]}" -- "$cur")
    fi
}

__kafkactl_handle_reply()
{
    __kafkactl_debug "${FUNCNAME[0]}"
    local comp
    case $cur in
        -*)
            if [[ $(type -t compopt) = "builtin" ]]; then
                compopt -o nospace
            fi
            local allflags
            if [ ${#must_have_one_flag[@]} -ne 0 ]; then
                allflags=("${must_have_one_flag[@]}")
            else
                allflags=("${flags[*]} ${two_word_flags[*]}")
            fi
            while IFS='' read -r comp; do
                COMPREPLY+=("$comp")
            done < <(compgen -W "${allflags[*]}" -- "$cur")
            if [[ $(type -t compopt) = "builtin" ]]; then
                [[ "${COMPREPLY[0]}" == *= ]] || compopt +o nospace
            fi

            # complete after --flag=abc
            if [[ $cur == *=* ]]; then
                if [[ $(type -t compopt) = "builtin" ]]; then
                    compopt +o nospace
                fi

                local index flag
                flag="${cur%=*}"
                __kafkactl_index_of_word "${flag}" "${flags_with_completion[@]}"
                COMPREPLY=()
                if [[ ${index} -ge 0 ]]; then
                    PREFIX=""
                    cur="${cur#*=}"
                    ${flags_completion[${index}]}
                    if [ -n "${ZSH_VERSION}" ]; then
                        # zsh completion needs --flag= prefix
                        eval "COMPREPLY=( \"\${COMPREPLY[@]/#/${flag}=}\" )"
                    fi
                fi
            fi
            return 0;
            ;;
    esac

    # check if we are handling a flag with special work handling
    local index
    __kafkactl_index_of_word "${prev}" "${flags_with_completion[@]}"
    if [[ ${index} -ge 0 ]]; then
        ${flags_completion[${index}]}
        return
    fi

    # we are parsing a flag and don't have a special handler, no completion
    if [[ ${cur} != "${words[cword]}" ]]; then
        return
    fi

    local completions
    completions=("${commands[@]}")
    if [[ ${#must_have_one_noun[@]} -ne 0 ]]; then
        completions+=("${must_have_one_noun[@]}")
    elif [[ -n "${has_completion_function}" ]]; then
        # if a go completion function is provided, defer to that function
        __kafkactl_handle_go_custom_completion
    fi
    if [[ ${#must_have_one_flag[@]} -ne 0 ]]; then
        completions+=("${must_have_one_flag[@]}")
    fi
    while IFS='' read -r comp; do
        COMPREPLY+=("$comp")
    done < <(compgen -W "${completions[*]}" -- "$cur")

    if [[ ${#COMPREPLY[@]} -eq 0 && ${#noun_aliases[@]} -gt 0 && ${#must_have_one_noun[@]} -ne 0 ]]; then
        while IFS='' read -r comp; do
            COMPREPLY+=("$comp")
        done < <(compgen -W "${noun_aliases[*]}" -- "$cur")
    fi

    if [[ ${#COMPREPLY[@]} -eq 0 ]]; then
		if declare -F __kafkactl_custom_func >/dev/null; then
			# try command name qualified custom func
			__kafkactl_custom_func
		else
			# otherwise fall back to unqualified for compatibility
			declare -F __custom_func >/dev/null && __custom_func
		fi
    fi

    # available in bash-completion >= 2, not always present on macOS
    if declare -F __ltrim_colon_completions >/dev/null; then
        __ltrim_colon_completions "$cur"
    fi

    # If there is only 1 completion and it is a flag with an = it will be completed
    # but we don't want a space after the =
    if [[ "${#COMPREPLY[@]}" -eq "1" ]] && [[ $(type -t compopt) = "builtin" ]] && [[ "${COMPREPLY[0]}" == --*= ]]; then
       compopt -o nospace
    fi
}

# The arguments should be in the form "ext1|ext2|extn"
__kafkactl_handle_filename_extension_flag()
{
    local ext="$1"
    _filedir "@(${ext})"
}

__kafkactl_handle_subdirs_in_dir_flag()
{
    local dir="$1"
    pushd "${dir}" >/dev/null 2>&1 && _filedir -d && popd >/dev/null 2>&1 || return
}

__kafkactl_handle_flag()
{
    __kafkactl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    # if a command required a flag, and we found it, unset must_have_one_flag()
    local flagname=${words[c]}
    local flagvalue
    # if the word contained an =
    if [[ ${words[c]} == *"="* ]]; then
        flagvalue=${flagname#*=} # take in as flagvalue after the =
        flagname=${flagname%=*} # strip everything after the =
        flagname="${flagname}=" # but put the = back
    fi
    __kafkactl_debug "${FUNCNAME[0]}: looking for ${flagname}"
    if __kafkactl_contains_word "${flagname}" "${must_have_one_flag[@]}"; then
        must_have_one_flag=()
    fi

    # if you set a flag which only applies to this command, don't show subcommands
    if __kafkactl_contains_word "${flagname}" "${local_nonpersistent_flags[@]}"; then
      commands=()
    fi

    # keep flag value with flagname as flaghash
    # flaghash variable is an associative array which is only supported in bash > 3.
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        if [ -n "${flagvalue}" ] ; then
            flaghash[${flagname}]=${flagvalue}
        elif [ -n "${words[ $((c+1)) ]}" ] ; then
            flaghash[${flagname}]=${words[ $((c+1)) ]}
        else
            flaghash[${flagname}]="true" # pad "true" for bool flag
        fi
    fi

    # skip the argument to a two word flag
    if [[ ${words[c]} != *"="* ]] && __kafkactl_contains_word "${words[c]}" "${two_word_flags[@]}"; then
			  __kafkactl_debug "${FUNCNAME[0]}: found a flag ${words[c]}, skip the next argument"
        c=$((c+1))
        # if we are looking for a flags value, don't show commands
        if [[ $c -eq $cword ]]; then
            commands=()
        fi
    fi

    c=$((c+1))

}

__kafkactl_handle_noun()
{
    __kafkactl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    if __kafkactl_contains_word "${words[c]}" "${must_have_one_noun[@]}"; then
        must_have_one_noun=()
    elif __kafkactl_contains_word "${words[c]}" "${noun_aliases[@]}"; then
        must_have_one_noun=()
    fi

    nouns+=("${words[c]}")
    c=$((c+1))
}

__kafkactl_handle_command()
{
    __kafkactl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"

    local next_command
    if [[ -n ${last_command} ]]; then
        next_command="_${last_command}_${words[c]//:/__}"
    else
        if [[ $c -eq 0 ]]; then
            next_command="_kafkactl_root_command"
        else
            next_command="_${words[c]//:/__}"
        fi
    fi
    c=$((c+1))
    __kafkactl_debug "${FUNCNAME[0]}: looking for ${next_command}"
    declare -F "$next_command" >/dev/null && $next_command
}

__kafkactl_handle_word()
{
    if [[ $c -ge $cword ]]; then
        __kafkactl_handle_reply
        return
    fi
    __kafkactl_debug "${FUNCNAME[0]}: c is $c words[c] is ${words[c]}"
    if [[ "${words[c]}" == -* ]]; then
        __kafkactl_handle_flag
    elif __kafkactl_contains_word "${words[c]}" "${commands[@]}"; then
        __kafkactl_handle_command
    elif [[ $c -eq 0 ]]; then
        __kafkactl_handle_command
    elif __kafkactl_contains_word "${words[c]}" "${command_aliases[@]}"; then
        # aliashash variable is an associative array which is only supported in bash > 3.
        if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
            words[c]=${aliashash[${words[c]}]}
            __kafkactl_handle_command
        else
            __kafkactl_handle_noun
        fi
    else
        __kafkactl_handle_noun
    fi
    __kafkactl_handle_word
}

_kafkactl_alter_topic()
{
    last_command="kafkactl_alter_topic"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config=")
    flags+=("--partitions=")
    two_word_flags+=("--partitions")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partitions=")
    flags+=("--validate-only")
    flags+=("-v")
    local_nonpersistent_flags+=("--validate-only")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_alter()
{
    last_command="kafkactl_alter"

    command_aliases=()

    commands=()
    commands+=("topic")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_attach()
{
    last_command="kafkactl_attach"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_completion()
{
    last_command="kafkactl_completion"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--help")
    flags+=("-h")
    local_nonpersistent_flags+=("--help")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    must_have_one_noun+=("bash")
    must_have_one_noun+=("fish")
    must_have_one_noun+=("powershell")
    must_have_one_noun+=("zsh")
    noun_aliases=()
}

_kafkactl_config_current-context()
{
    last_command="kafkactl_config_current-context"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_config_get-contexts()
{
    last_command="kafkactl_config_get-contexts"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_config_use-context()
{
    last_command="kafkactl_config_use-context"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_config_view()
{
    last_command="kafkactl_config_view"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_config()
{
    last_command="kafkactl_config"

    command_aliases=()

    commands=()
    commands+=("current-context")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("currentContext")
        aliashash["currentContext"]="current-context"
    fi
    commands+=("get-contexts")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("getContexts")
        aliashash["getContexts"]="get-contexts"
    fi
    commands+=("use-context")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("useContext")
        aliashash["useContext"]="use-context"
    fi
    commands+=("view")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_consume()
{
    last_command="kafkactl_consume"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--exit")
    flags+=("-e")
    local_nonpersistent_flags+=("--exit")
    flags+=("--from-beginning")
    flags+=("-b")
    local_nonpersistent_flags+=("--from-beginning")
    flags+=("--key-encoding=")
    two_word_flags+=("--key-encoding")
    local_nonpersistent_flags+=("--key-encoding=")
    flags+=("--offset=")
    two_word_flags+=("--offset")
    local_nonpersistent_flags+=("--offset=")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--partitions=")
    two_word_flags+=("--partitions")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partitions=")
    flags+=("--print-headers")
    local_nonpersistent_flags+=("--print-headers")
    flags+=("--print-keys")
    flags+=("-k")
    local_nonpersistent_flags+=("--print-keys")
    flags+=("--print-schema")
    flags+=("-a")
    local_nonpersistent_flags+=("--print-schema")
    flags+=("--print-timestamps")
    flags+=("-t")
    local_nonpersistent_flags+=("--print-timestamps")
    flags+=("--separator=")
    two_word_flags+=("--separator")
    two_word_flags+=("-s")
    local_nonpersistent_flags+=("--separator=")
    flags+=("--tail=")
    two_word_flags+=("--tail")
    local_nonpersistent_flags+=("--tail=")
    flags+=("--value-encoding=")
    two_word_flags+=("--value-encoding")
    local_nonpersistent_flags+=("--value-encoding=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_create_consumer-group()
{
    last_command="kafkactl_create_consumer-group"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--newest")
    local_nonpersistent_flags+=("--newest")
    flags+=("--offset=")
    two_word_flags+=("--offset")
    local_nonpersistent_flags+=("--offset=")
    flags+=("--oldest")
    local_nonpersistent_flags+=("--oldest")
    flags+=("--partition=")
    two_word_flags+=("--partition")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partition=")
    flags+=("--topic=")
    two_word_flags+=("--topic")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--topic=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_create_topic()
{
    last_command="kafkactl_create_topic"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config=")
    two_word_flags+=("--config")
    two_word_flags+=("-c")
    local_nonpersistent_flags+=("--config=")
    flags+=("--partitions=")
    two_word_flags+=("--partitions")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partitions=")
    flags+=("--replication-factor=")
    two_word_flags+=("--replication-factor")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--replication-factor=")
    flags+=("--validate-only")
    flags+=("-v")
    local_nonpersistent_flags+=("--validate-only")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_create()
{
    last_command="kafkactl_create"

    command_aliases=()

    commands=()
    commands+=("consumer-group")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cg")
        aliashash["cg"]="consumer-group"
    fi
    commands+=("topic")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_delete_topic()
{
    last_command="kafkactl_delete_topic"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_delete()
{
    last_command="kafkactl_delete"

    command_aliases=()

    commands=()
    commands+=("topic")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_describe_consumer-group()
{
    last_command="kafkactl_describe_consumer-group"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--only-with-lag")
    flags+=("-l")
    local_nonpersistent_flags+=("--only-with-lag")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--print-members")
    flags+=("-m")
    local_nonpersistent_flags+=("--print-members")
    flags+=("--print-topics")
    flags+=("-T")
    local_nonpersistent_flags+=("--print-topics")
    flags+=("--topic=")
    two_word_flags+=("--topic")
    flags_with_completion+=("--topic")
    flags_completion+=("__kafkactl_handle_go_custom_completion")
    two_word_flags+=("-t")
    flags_with_completion+=("-t")
    flags_completion+=("__kafkactl_handle_go_custom_completion")
    local_nonpersistent_flags+=("--topic=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_describe_topic()
{
    last_command="kafkactl_describe_topic"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--print-configs")
    flags+=("-c")
    local_nonpersistent_flags+=("--print-configs")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_describe()
{
    last_command="kafkactl_describe"

    command_aliases=()

    commands=()
    commands+=("consumer-group")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cg")
        aliashash["cg"]="consumer-group"
    fi
    commands+=("topic")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_get_consumer-groups()
{
    last_command="kafkactl_get_consumer-groups"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--topic=")
    two_word_flags+=("--topic")
    flags_with_completion+=("--topic")
    flags_completion+=("__kafkactl_handle_go_custom_completion")
    two_word_flags+=("-t")
    flags_with_completion+=("-t")
    flags_completion+=("__kafkactl_handle_go_custom_completion")
    local_nonpersistent_flags+=("--topic=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_get_topics()
{
    last_command="kafkactl_get_topics"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_get()
{
    last_command="kafkactl_get"

    command_aliases=()

    commands=()
    commands+=("consumer-groups")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cg")
        aliashash["cg"]="consumer-groups"
    fi
    commands+=("topics")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_help()
{
    last_command="kafkactl_help"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_produce()
{
    last_command="kafkactl_produce"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--file=")
    two_word_flags+=("--file")
    two_word_flags+=("-f")
    local_nonpersistent_flags+=("--file=")
    flags+=("--header=")
    two_word_flags+=("--header")
    two_word_flags+=("-H")
    local_nonpersistent_flags+=("--header=")
    flags+=("--key=")
    two_word_flags+=("--key")
    two_word_flags+=("-k")
    local_nonpersistent_flags+=("--key=")
    flags+=("--key-encoding=")
    two_word_flags+=("--key-encoding")
    local_nonpersistent_flags+=("--key-encoding=")
    flags+=("--key-schema-version=")
    two_word_flags+=("--key-schema-version")
    two_word_flags+=("-K")
    local_nonpersistent_flags+=("--key-schema-version=")
    flags+=("--lineSeparator=")
    two_word_flags+=("--lineSeparator")
    two_word_flags+=("-L")
    local_nonpersistent_flags+=("--lineSeparator=")
    flags+=("--partition=")
    two_word_flags+=("--partition")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partition=")
    flags+=("--partitioner=")
    two_word_flags+=("--partitioner")
    two_word_flags+=("-P")
    local_nonpersistent_flags+=("--partitioner=")
    flags+=("--rate=")
    two_word_flags+=("--rate")
    two_word_flags+=("-r")
    local_nonpersistent_flags+=("--rate=")
    flags+=("--separator=")
    two_word_flags+=("--separator")
    two_word_flags+=("-S")
    local_nonpersistent_flags+=("--separator=")
    flags+=("--silent")
    flags+=("-s")
    local_nonpersistent_flags+=("--silent")
    flags+=("--value=")
    two_word_flags+=("--value")
    two_word_flags+=("-v")
    local_nonpersistent_flags+=("--value=")
    flags+=("--value-encoding=")
    two_word_flags+=("--value-encoding")
    local_nonpersistent_flags+=("--value-encoding=")
    flags+=("--value-schema-version=")
    two_word_flags+=("--value-schema-version")
    two_word_flags+=("-i")
    local_nonpersistent_flags+=("--value-schema-version=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_reset_consumer-group-offset()
{
    last_command="kafkactl_reset_consumer-group-offset"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--execute")
    flags+=("-e")
    local_nonpersistent_flags+=("--execute")
    flags+=("--newest")
    local_nonpersistent_flags+=("--newest")
    flags+=("--offset=")
    two_word_flags+=("--offset")
    local_nonpersistent_flags+=("--offset=")
    flags+=("--oldest")
    local_nonpersistent_flags+=("--oldest")
    flags+=("--output=")
    two_word_flags+=("--output")
    two_word_flags+=("-o")
    local_nonpersistent_flags+=("--output=")
    flags+=("--partition=")
    two_word_flags+=("--partition")
    two_word_flags+=("-p")
    local_nonpersistent_flags+=("--partition=")
    flags+=("--topic=")
    two_word_flags+=("--topic")
    two_word_flags+=("-t")
    local_nonpersistent_flags+=("--topic=")
    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    has_completion_function=1
    noun_aliases=()
}

_kafkactl_reset()
{
    last_command="kafkactl_reset"

    command_aliases=()

    commands=()
    commands+=("consumer-group-offset")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("cgo")
        aliashash["cgo"]="consumer-group-offset"
        command_aliases+=("offset")
        aliashash["offset"]="consumer-group-offset"
    fi

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_version()
{
    last_command="kafkactl_version"

    command_aliases=()

    commands=()

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

_kafkactl_root_command()
{
    last_command="kafkactl"

    command_aliases=()

    commands=()
    commands+=("alter")
    commands+=("attach")
    commands+=("completion")
    commands+=("config")
    commands+=("consume")
    commands+=("create")
    commands+=("delete")
    commands+=("describe")
    commands+=("get")
    if [[ -z "${BASH_VERSION}" || "${BASH_VERSINFO[0]}" -gt 3 ]]; then
        command_aliases+=("list")
        aliashash["list"]="get"
    fi
    commands+=("help")
    commands+=("produce")
    commands+=("reset")
    commands+=("version")

    flags=()
    two_word_flags=()
    local_nonpersistent_flags=()
    flags_with_completion=()
    flags_completion=()

    flags+=("--config-file=")
    two_word_flags+=("--config-file")
    two_word_flags+=("-C")
    flags+=("--verbose")
    flags+=("-V")

    must_have_one_flag=()
    must_have_one_noun=()
    noun_aliases=()
}

__start_kafkactl()
{
    local cur prev words cword
    declare -A flaghash 2>/dev/null || :
    declare -A aliashash 2>/dev/null || :
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -s || return
    else
        __kafkactl_init_completion -n "=" || return
    fi

    local c=0
    local flags=()
    local two_word_flags=()
    local local_nonpersistent_flags=()
    local flags_with_completion=()
    local flags_completion=()
    local commands=("kafkactl")
    local must_have_one_flag=()
    local must_have_one_noun=()
    local has_completion_function
    local last_command
    local nouns=()

    __kafkactl_handle_word
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __start_kafkactl kafkactl
else
    complete -o default -o nospace -F __start_kafkactl kafkactl
fi

# ex: ts=4 sw=4 et filetype=sh
