module.exports = {
	rules: {
		"commit-msg": [2, "always"],
	},
	plugins: [
		{
			rules: {
				"commit-msg": ({ header }) => {
					const commitMessage = header.trim();
					if (commitMessage.length <= 6) {
						return [false, "Commit message must be over 6 characters"];
					}
					if (commitMessage.length > 79) {
						return [false, "Commit message must not exceed 79 characters"];
					}
					return [true];
				},
			},
		},
	],
};
