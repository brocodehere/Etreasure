/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{astro,html,js,jsx,md,mdx,svelte,ts,tsx,vue}'],
  theme: {
    extend: {
      colors: {
        maroon: "#6D0F19",
        gold: "#D4A857", 
        cream: "#F7F3EB",
        indigo: "#1A1E3A",
        forest: "#0B3D2E",
        dark: "#161616"
      },
      fontFamily: {
        'playfair': ['Playfair Display', 'serif'],
        'inter': ['Inter', 'sans-serif'], 
        'poppins': ['Poppins', 'sans-serif']
      },
      lineHeight: {
        'heading': '1.3',
        'body': '1.6'
      },
      letterSpacing: {
        'wide': '0.05em',
        'wider': '0.1em', 
        'widest': '0.2em'
      },
      animation: {
        'fade-in': 'fadeIn 0.8s ease-out',
        'fade-up': 'fadeUp 0.8s ease-out',
        'slide-in': 'slideIn 0.6s ease-out',
        'zoom-in': 'zoomIn 0.5s ease-out',
        'gold-shine': 'goldShine 2s ease-in-out infinite'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(20px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideIn: {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(0)' }
        },
        zoomIn: {
          '0%': { transform: 'scale(0.95)' },
          '100%': { transform: 'scale(1)' }
        },
        goldShine: {
          '0%': { boxShadow: '0 0 0 0 rgba(212, 168, 87, 0.4)' },
          '50%': { boxShadow: '0 0 0 10px rgba(212, 168, 87, 0)' },
          '100%': { boxShadow: '0 0 0 0 rgba(212, 168, 87, 0)' }
        }
      },
      boxShadow: {
        'gold': '0 0 20px rgba(212, 168, 87, 0.3)',
        'gold-hover': '0 0 30px rgba(212, 168, 87, 0.5)',
        'card': '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)'
      }
    },
  },
  plugins: [],
}
