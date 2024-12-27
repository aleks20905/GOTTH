const colors = require('tailwindcss/colors');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    'internal/templates/*.templ',
  ],
  theme: {
    container: {
      center: true,
      padding: {
        DEFAULT: "1rem",
        mobile: "2rem",
        tablet: "4rem",
        desktop: "5rem",
      },
    },
    extend: {
      colors: {
        primary: colors.gray[400], // Lighter gray for primary accents
        secondary: colors.amber[400], // Amber accent for hover and links
        neutral: 'rgb(24, 26, 27)', // Darker gray background (RGB(24, 26, 27))
        background: 'rgb(24, 26, 27)', // Darker gray background (RGB(24, 26, 27))
        text: colors.white, // White text for contrast on dark background
        navbarBackground: 'rgb(20, 22, 24)', // Darker shade for navbar (RGB(20, 22, 24))
        navbarText: colors.gray[200], // Light gray text for the navbar
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ]
}

