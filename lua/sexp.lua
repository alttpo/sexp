do
    local _b = string.byte
    local _ch_digit_0 = _b('0')
    local _ch_digit_9 = _b('9')
    local _ch_upper_A = _b('A')
    local _ch_upper_Z = _b('Z')
    local _ch_lower_A = _b('a')
    local _ch_lower_Z = _b('z')
    local _ch_cr = _b('\r')
    local _ch_lf = _b('\n')
    local _ch_dot = _b('.')
    local _ch_slash = _b('/')
    local _ch_under = _b('_')
    local _ch_colon = _b(':')
    local _ch_asterisk = _b('*')
    local _ch_plus = _b('+')
    local _ch_minus = _b('-')
    local _ch_equals = _b('=')
    local _ch_hash = _b('#')
    local _ch_space = _b(' ')
    local _ch_tab = _b('\t')
    local _ch_vtab = _b('\v')
    local _ch_ff = _b('\f')
    local _ch_paren_close = _b(')')
    local _ch_paren_open = _b('(')

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
        local function serr(str)
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
                return node(), i, serr('not ascii')
            elseif r == _ch_cr or r == _ch_lf then
                return node(), i, serr('newlines not allowed')
            end

            if is_token_remainder(r) then
                last = i

                -- safe to advance:
                i = i + 1
            else
                -- hit something else:
                return node(), i, nil
            end
        end

        -- all we can say here is we reached the end of the string:
        return node(), i, serr('eof')
    end

    local parse_node
    -- s: string:  s-expression to parse
    -- i: integer: starting index
    -- returns:
    --   n:  table:   node or nil
    --   i:  integer: ending index
    --   err: table:  error or nil
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
                return n, i, nil
            end

            n.list[#n.list+1] = child
        end

        return n, i, serr('eof')
    end

    local hexmap = {}
    do
        hexmap[_b('0')] = 0x0
        hexmap[_b('1')] = 0x1
        hexmap[_b('2')] = 0x2
        hexmap[_b('3')] = 0x3
        hexmap[_b('4')] = 0x4
        hexmap[_b('5')] = 0x5
        hexmap[_b('6')] = 0x6
        hexmap[_b('7')] = 0x7
        hexmap[_b('8')] = 0x8
        hexmap[_b('9')] = 0x9
        hexmap[_b('a')] = 0xa
        hexmap[_b('b')] = 0xb
        hexmap[_b('c')] = 0xc
        hexmap[_b('d')] = 0xd
        hexmap[_b('e')] = 0xe
        hexmap[_b('f')] = 0xf
        hexmap[_b('A')] = 0xA
        hexmap[_b('B')] = 0xB
        hexmap[_b('C')] = 0xC
        hexmap[_b('D')] = 0xD
        hexmap[_b('E')] = 0xE
        hexmap[_b('F')] = 0xF
    end
    local function parse_hex(s, i)
        local function serr(str)
            return { err = str; i = i }
        end

        local n = { hex = "" }
        local c = {}
        local d, e = 1, 0
        local function node()
            n.hex = string.char(unpack(c))
            return n
        end

        while i <= #s do
            local r = s:byte(i)
            -- validate character range:
            if r > 128 then
                return node(), i, serr('not ascii')
            elseif r == _ch_cr or r == _ch_lf then
                return node(), i, serr('newlines not allowed')
            end

            -- safe to advance:
            i = i + 1

            h = hexmap[r]
            if is_space(r) then
                -- skip whitespace
            elseif r == _ch_hash then
                -- end of hex:
                return node(), i, nil
            elseif h ~= nil then
                -- add hex char:
                e = e + 1
                if e == 2 then
                    c[d] = c[d] + h
                    d = d + 1
                    e = 0
                else
                    c[d] = h * 16
                end
            else
                return node(), i, serr('unexpected character')
            end
        end

        return node(), i, serr('eof')
    end

    -- s: string:  s-expression to parse
    -- i: integer: starting index
    -- returns:
    --   n:  table:   node or nil
    --   i:  integer: ending index
    --   eol: bool:   end of list
    --   err: table:  error or nil
    parse_node = function(s, i)
        local function serr(str)
            return { err = str; i = i }
        end

        local n, err

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
                return nil, i, true, nil
            elseif r == _ch_paren_open then
                -- start of list
                n, i, err = parse_list(s, i)
                return n, i, false, err
            elseif r == _ch_hash then
                n, i, err = parse_hex(s, i)
                return n, i, false, err
            elseif is_token_start(r) then
                n, i, err = parse_token(s, i - 1)
                return n, i, false, err
            else
                return nil, i, false, serr('unexpected character')
            end
        end

        return nil, i, false, serr('eof')
    end

    -------------
    -- decoder --
    -------------
    function sexp_decode(s)
        if type(s) ~= 'string' then
            return nil, 0, { err = 'expected string argument', i = 0 }
        end

        local n, i, eol, err = parse_node(s, 1)
        if eol then
            return n, i, { err = 'unexpected end of list', i = i }
        end
        return n, i, err
    end

    local function sexp_encode_node(n, sb)
        local err
        if type(n.list) == 'table' then
            sb[#sb+1] = '('
            for i, c in ipairs(n.list) do
                err = sexp_encode_node(c, sb)
                if err then
                    return err
                end
                if i < #n.list then
                    sb[#sb+1] = ' '
                end
            end
            sb[#sb+1] = ')'
            return nil
        elseif type(n.hex) == 'string' then
            sb[#sb+1] = '#'
            for i=1,#n.hex do
                sb[#sb+1] = string.format("%02x", n.hex:byte(i))
            end
            sb[#sb+1] = '#'
            return nil
        elseif type(n.token) == 'string' then
            sb[#sb+1] = n.token
            return nil
        else
            return { err = 'unknown node type' }
        end
    end

    -------------
    -- encoder --
    -------------
    function sexp_encode(n)
        local sb = {}
        local err = sexp_encode_node(n, sb)
        return table.concat(sb), err
    end
end