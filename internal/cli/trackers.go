package cli

// trackersHelp is the one answer to "where does this get filed", shared by
// every command that files something, so they cannot drift into separate
// answers.
//
// It exists because each tracker used to describe only itself. The Workbench
// help said it was not question/request/task, issue-report never named the
// Workbench, and both call a defect a bug. Agents given only that text, and
// asked where eight ordinary situations go, filed a customer-side blocker in
// the repository where the customer never sees it.
const trackersHelp = `WHERE A THING IS FILED - decided by who has to act on it

    the workspace's Workbench - the customer reads it
        asgard-cli workbench create
        when the customer's side has to see, answer, supply or decide it,
        or something is wrong in what is live

    this repository's records - the engagement's own state
        asgard-cli question add, asgard-cli request add, asgard-cli task add
        when whoever builds next needs it: the spec, the design, an open
        decision

    upstream - a feedback only the tool's maintainers read
        asgard-cli issue-report --new, then --send
        when the makers of asgard-cli, or of the platform behind it, have
        to fix something

  - **A failure of asgard-cli is never a Workbench bug**, not even when it
    happened while working on an issue. The Workbench's bug type is what is
    wrong in the customer's deployment, and the customer reads it. Report the
    tool upstream; on the issue, say at most that the work is blocked. A
    command that failed, or blamed the wrong thing, is the tool's - the
    maintainers route it on if the platform is at fault.
  - **Nothing about the customer goes upstream**: not their systems, their
    deployment, their data or their people - a report leaves the engagement
    for a third party's service, read by people who are not the customer's. Their
    ERP failing, what they still owe, the bot they run answering wrongly: each
    is a Workbench issue (a question while it blocks the build, a bug once it
    is live).
  - **One thing can need two places.** What the customer asked for is a
    request spec here, because the spec is what gets built from; if the
    customer follows it on the Workbench as well - ask, rather than assume -
    write the ISS-N in the spec so each is found from the other. The same
    goes for a question the customer has to answer that also shapes the
    design.
  - **In the Workbench assistant's sandbox** what the member has to keep
    track of goes on the Workbench: a repository record reaches nobody until
    somebody with administration pushes it. A gap in the tool is still
    reported with asgard-cli issue-report, as anywhere else - and the member
    may be the customer's own staff, so the rule about their content holds
    there as much as here.`
