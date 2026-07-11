# Your first PR in 5 minutes

This guide is for people who have never contributed to open source before. If anything feels confusing, open a discussion and we will fix the guide.

## What you will do

You will add one small improvement to Relay and open a pull request.

## Before you start

You need:
- Git
- Go 1.22 or newer

## Steps

### 1. Fork and clone

Click **Fork** on https://github.com/valtors/relay, then:

```bash
git clone https://github.com/YOUR_USERNAME/relay
cd relay
```

### 2. Pick an issue

Find an issue labeled [`good first issue`](https://github.com/valtors/relay/labels/good%20first%20issue). Comment on it so others know you are working on it.

If you do not want to pick an issue, add one small tool. See [docs/ADDING_A_TOOL.md](ADDING_A_TOOL.md).

### 3. Make the change

Edit one file. Keep it small.

### 4. Run the checks

```bash
gofmt -w .
go vet ./...
go test ./...
```

Make sure all three pass.

### 5. Commit and push

```bash
git add .
git commit -m "fix: short lowercase message"
git push origin main
```

### 6. Open a PR

Go to https://github.com/valtors/relay/pulls and click **New pull request**. Fill in the template. Link the issue with `Closes #123`.

## What happens next

A maintainer will review within 3 business days. We may ask for small changes. That is normal.

## If you get stuck

Open a [Discussion](https://github.com/valtors/relay/discussions) with the title "I need help with my first PR". We will help.
