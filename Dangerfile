# Check if Merge Request is a Work in progress and display a warning
is_draft = gitlab.pr_title.include? "Draft:"
warn("MR is classed as Work in Progress") if is_draft

# Check validity of MR title. Jira ticket is mandatory for FEAT/FIX MR, optional for CHORE:
# ✅ feat(DEVOPS-123): description
# ❌ feat: description
# ❌ feat(DEVOPS-555): description
# ✅ fix(DEVOPS-340): description
# ❌ fix: description
# ✅ chore(scope): description
# ❌ Fix/scope/description
# ❌ CHORE
# ❌ My merge request title is not compliant
# More details: https://tamarapay.atlassian.net/wiki/spaces/MVP/pages/2220392518/GitLab+Handbook+Usage
valid_mr_title = !!(gitlab.mr_title =~ /(^((build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([^)]+\))?(!)?(:(.*\s*)*)))/)
if !valid_mr_title && !is_draft
  fail('Please make sure your MR title respects the correct commit format: `feat|fix|perf|refactor`(`JIRA-TICKET`): `{description}` or `ci|docs|chore`(`scope`): `{description}`. More details: https://tamarapay.atlassian.net/wiki/spaces/MVP/pages/2220392518/GitLab+Handbook+Usage .', sticky: true)
end

# Warn when a Merge Request is very big to sensibilize developers
big_mr_warning = "Merge request contains a lot of changes, please consider splitting it. Big merge requests can delay review and approvals"
bitMRLimit = (git.modified_files) ? 1200 : 700
warn(big_mr_warning) if git.lines_of_code > bitMRLimit

# Warn when a Merge Request has too many commits
many_commits_mr_warning = "Merge request includes more than 50 commits. Please rebase/squash these commits into a smaller number of commits."
warn(many_commits_mr_warning) if git.commits.length > 50

unless gitlab.mr_json["assignee"]
  warn "This merge request does not have any assignee yet. Setting an assignee clarifies who needs to take action on the merge request at any given time."
end


# Warn when there is a small PR (testing purpose)
markdown("✅ Yay, a small MR  🎉 Keep it up !") if git.lines_of_code < 100
# frozen_string_literal: true

markdown("/assign_reviewer @tamara-backend/tech-org/chapter/go")

DB_MESSAGE = <<~MSG
This merge request requires a database review. To make sure these changes are reviewed, take the following steps:

1. Ensure the merge request has ~database and ~"database::review pending" labels.
   If the merge request modifies database files, Danger will do this for you.
2. Assign and mention the database engineering @tamara-backend/tech-org/data/data-engineering as reviewer.
MSG

DB_FILES_MESSAGE = <<~MSG
The following files require a review from the Database team:
MSG

migration_created = !git.added_files.grep(%r{\App/src/Tamara/Infrastructure/Migrations/}).empty?

if migration_created
  warn format(DB_MESSAGE, migrations: 'migrations')
  markdown("/assign_reviewer @tamara-backend/tech-org/data/data-engineering")
end
