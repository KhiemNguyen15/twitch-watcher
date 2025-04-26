module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "scope-enum": [
      2,
      "always",
      [
        "stream-poller",
        "webhook-sender",
        "user-service",
        "frontend",
        "infra",
        "auth",
      ],
    ],
    "subject-case": [
      2,
      "always",
      ["sentence-case", "start-case", "pascal-case", "upper-case"],
    ],
  },
};
