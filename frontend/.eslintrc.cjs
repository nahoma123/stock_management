module.exports = {
    root: true,
    env: { browser: true, es2021: true },
    extends: ['eslint:recommended', 'plugin:react/recommended', 'plugin:react-hooks/recommended'],
    ignorePatterns: ['dist'],
    parserOptions: { ecmaVersion: 'latest', sourceType: 'module', ecmaFeatures: { jsx: true } },
    plugins: ['react-refresh'],
    settings: { react: { version: 'detect' } },
    rules: {
        'react/prop-types': 'off',
        'react/react-in-jsx-scope': 'off',
        'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
    },
};
