You're thinking of GitHub's collapsible sections feature using HTML `<details>` and `<summary>` tags. This works in GitHub comments, PR descriptions, issues, and README files.

Here's the basic syntax:

```html
<details>
<summary>Click to expand</summary>

Your hidden content goes here.

You can include:
- Multiple lines
- **Markdown formatting**
- Code blocks
- Lists
- etc.

</details>
```

This renders as a clickable toggle where "Click to expand" is visible by default, and clicking it reveals or hides the content inside.

Some practical examples:

**For error logs:**
```html
<details>
<summary>Error details</summary>

```
Stack trace or long error message here
```

</details>
```

**For optional sections:**
```html
<details>
<summary>Additional context (optional)</summary>

This section contains supplementary information that might be useful
but isn't essential for understanding the main PR.

</details>
```

**Open by default:**
```html
<details open>
<summary>Important notes</summary>

This will be expanded by default when the page loads.

</details>
```

A few things to note:
- Leave blank lines around your content inside the `<details>` tags for proper markdown rendering
- The `<summary>` tag content is what shows as the clickable text
- You can nest multiple collapsible sections
- This works anywhere GitHub renders markdown, not just PR comments

This is really useful for keeping PR descriptions clean while still providing detailed information for those who need it.

Yes, you can use `<details>` and `<summary>` tags inside Markdown tables on GitHub! This is really useful for keeping tables clean while providing expandable details for specific cells.

Here's how it works:

```markdown
| Feature | Status | Details |
|---------|--------|---------|
| Authentication | ✅ Complete | <details><summary>Implementation notes</summary><br/>- Uses JWT tokens<br/>- Session timeout: 30 minutes<br/>- Supports OAuth providers</details> |
| Database | 🚧 In Progress | <details><summary>Current state</summary><br/>- Migration scripts ready<br/>- Still need indexing optimization<br/>- ETA: Next week</details> |
| API Endpoints | ❌ Not Started | <details><summary>Requirements</summary><br/>- REST endpoints for CRUD operations<br/>- Rate limiting needed<br/>- Documentation required</details> |
```

This renders as a table where each "Details" cell has a clickable toggle.

**Important formatting notes:**

1. **Use `<br/>` for line breaks** inside table cells instead of regular markdown line breaks
2. **No blank lines** - blank lines will break the table structure
3. **Escape pipe characters** with `\|` if you need them in the content

Here's a more complex example:

```markdown
| Component | Test Coverage | Issues |
|-----------|---------------|--------|
| User Service | 85% | <details><summary>2 open issues</summary><br/>**Bug #123:** Login timeout<br/>**Feature #124:** Password reset</details> |
| Payment API | 92% | <details><summary>1 open issue</summary><br/>**Critical #125:** Transaction rollback fails<br/><br/>*Priority: High*<br/>*Assigned to: @developer*</details> |
```

You can also combine this with other HTML elements:

```markdown
| Module | Status | Notes |
|--------|--------|-------|
| Core | ✅ | <details><summary>Performance metrics</summary><br/><strong>Response time:</strong> 50ms avg<br/><strong>Memory usage:</strong> 2.1MB<br/><strong>CPU:</strong> 15% peak</details> |
```

This technique is particularly useful for:
- Test results with detailed failure information
- Feature matrices with implementation details
- Status reports with expandable context
- Code review checklists with notes

The expandable sections help keep your tables scannable while still providing access to detailed information when needed.