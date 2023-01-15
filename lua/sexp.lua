do
    local _ch_digit_0 = string.byte('0')
    local _ch_digit_9 = string.byte('9')
    local _ch_upper_A = string.byte('A')
    local _ch_upper_Z = string.byte('Z')
    local _ch_lower_A = string.byte('a')
    local _ch_lower_Z = string.byte('z')
    local _ch_cr = string.byte('\r')
    local _ch_lf = string.byte('\n')
    local _ch_dot = string.byte('.')
    local _ch_slash = string.byte('/')
    local _ch_under = string.byte('_')
    local _ch_colon = string.byte(':')
    local _ch_asterisk = string.byte('*')
    local _ch_plus = string.byte('+')
    local _ch_minus = string.byte('-')
    local _ch_equals = string.byte('=')
    local _ch_space = string.byte(' ')
    local _ch_tab = string.byte('\t')
    local _ch_vtab = string.byte('\v')
    local _ch_ff = string.byte('\f')
    local _ch_paren_close = string.byte(')')
    local _ch_paren_open = string.byte('(')

    local function is_space(r)
        if r <= _ch_space or r == _ch_tab or r == _ch_vtab or r == _ch_ff then
            return true
        end
        return false
    end

    local function is_alpha(r)
        if r >= _ch_upper_A and r <= _ch_upper_Z then
            return true
        end
        if r >= _ch_lower_A and r <= _ch_lower_Z then
            return true
        end
        return false
    end

    local function is_digit(r)
        if r >= _ch_digit_0 and r <= _ch_digit_9 then
            return true
        end
        return false
    end

    local function is_graphic(r)
        return r == _ch_minus or
            r == _ch_dot or
            r == _ch_slash or
            r == _ch_under or
            r == _ch_colon or
            r == _ch_asterisk or
            r == _ch_plus or
            r == _ch_equals
    end

    local function is_token_start(r)
        return is_alpha(r) or is_graphic(r)
    end

    local function is_token_remainder(r)
        return is_alpha(r) or is_digit(r) or is_graphic(r)
    end

    local function parse_token(s, i)
        local function err(str)
            return { err = str; i = i }
        end

        -- temporary state:
        local start = i
        local last = i
        local function node()
            return {
                token = s:sub(start, last)
            }
        end

        while i <= #s do
            local r = s:byte(i)
            -- validate character range:
            if r > 128 then
                return node(), i, err('not ascii')
            elseif r == _ch_cr or r == _ch_lf then
                return node(), i, err('newlines not allowed')
            end

            if is_token_remainder(r) then
                last = i

                -- safe to advance:
                i = i + 1
            else
                -- hit something else:
                return node(), i, null
            end
        end

        -- all we can say here is we reached the end of the string:
        return node(), i, err('eof')
    end

    local parse_node
    -- s: string:  s-expression to parse
    -- i: integer: starting index
    -- returns:
    --   n:  table:   node or null
    --   i:  integer: ending index
    --   err: table:  error or null
    local function parse_list(s, i)
        local function serr(str)
            return { err = str; i = i }
        end

        local child, eol, err
        local n = {
            list = {}
        }
        while i <= #s do
            local r = s:byte(i)
            -- validate character range:
            if r > 128 then
                return n, i, serr('not ascii')
            elseif r == _ch_cr or r == _ch_lf then
                return n, i, serr('newlines not allowed')
            end

            child, i, eol, err = parse_node(s, i)
            if err then
                return n, i, err
            end
            if eol then
                return n, i, null
            end

            n.list[#n.list+1] = child
        end

        return n, i, serr('eof')
    end

    -- s: string:  s-expression to parse
    -- i: integer: starting index
    -- returns:
    --   n:  table:   node or null
    --   i:  integer: ending index
    --   eol: bool:   end of list
    --   err: table:  error or null
    parse_node = function(s, i)
        local function serr(str)
            return { err = str; i = i }
        end

        local n = {
            list = {}
        }

        while i <= #s do
            local r = s:byte(i)
            -- validate character range:
            if r > 128 then
                return n, i, false, serr('not ascii')
            elseif r == _ch_cr or r == _ch_lf then
                return n, i, false, serr('newlines not allowed')
            end

            -- safe to advance:
            i = i + 1

            if is_space(r) then
                -- skip whitespace
            elseif r == _ch_paren_close then
                -- end of list
                return null, i, true, null
            elseif r == _ch_paren_open then
                -- start of list
                return parse_list(s, i)
            elseif is_token_start(r) then
                return parse_token(s, i - 1)
            else
                return null, i, false, serr('unexpected character')
            end
        end

        return null, i, false, serr('eof')
    end

    -- global entry point for parsing
    function sexp_parse(s)
        local n, i, eol, err = parse_node(s, 1)
        return n, i, eol, err
    end
end
