-- Print-only callouts. A heading explicitly beginning "Sidebar:" opens a
-- sidebar through the next heading of the same or higher level. Ordinary
-- sections such as "Sidebar tabs", and headings inside code, are untouched.
function Blocks(blocks)
  local result = pandoc.List()
  local i = 1
  while i <= #blocks do
    local heading = blocks[i]
    if heading.t == 'Header' and
        pandoc.utils.stringify(heading.content):match('^Sidebar:%s') then
      local contents = pandoc.List({heading})
      heading.classes:insert('print-sidebar-title')
      local j = i + 1
      while j <= #blocks do
        local next_block = blocks[j]
        if next_block.t == 'Header' and next_block.level <= heading.level then
          break
        end
        contents:insert(next_block)
        j = j + 1
      end
      local attrs = {role = 'note'}
      if heading.identifier ~= '' then
        attrs['aria-labelledby'] = heading.identifier
      end
      result:insert(pandoc.Div(contents, pandoc.Attr('', {'print-sidebar'}, attrs)))
      i = j
    else
      result:insert(heading)
      i = i + 1
    end
  end
  return result
end
