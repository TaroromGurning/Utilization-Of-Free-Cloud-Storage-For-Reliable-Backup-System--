import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '@/stores/authStore.js'; // Store Otentikasi
import { useAppStore } from '@/stores/app.js' // ⭐ KOREKSI: Pastikan ini adalah path file Store Setup Anda

// ============================================
// IMPORT LAYOUT & PAGES
// ============================================
import AppLayout from '@/views/AppLayout.vue';
import Login from '@/views/Login.vue';
import SetupWizard from '@/views/SetupWizard.vue'; // ⭐ BARU: Import Halaman Setup
import Dashboard from '@/views/Dashboard.vue';
import Logs from '@/views/Logs.vue';
import ManualJobs from '@/views/ManualJobs.vue';
import Remotes from '@/views/Remotes.vue';
import ScheduledJobs from '@/views/ScheduledJobs.vue';

// ============================================
// ROUTE CONFIGURATION
// ============================================
const routes = [
    // ----------------------------------------
    // PUBLIC ROUTES (No Authentication)
    // ----------------------------------------
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { 
            requiresAuth: false,
            title: 'Login - G-Backup'
        }
    },
    // ⭐ BARU: SETUP WIZARD ROUTE
    {
        path: '/setup-wizard',
        name: 'SetupWizard',
        component: SetupWizard,
        meta: {
            requiresAuth: false,
            isSetupRoute: true, // Tag untuk membedakan rute Setup
            title: 'Setup Awal - G-Backup'
        }
    },
    
    // ----------------------------------------
    // PROTECTED ROUTES (With Authentication)
    // ----------------------------------------
    {
        path: '/',
        component: AppLayout,  // LAYOUT WRAPPER
        meta: { requiresAuth: true },
        children: [
            // Default redirect
            {
                path: '',
                redirect: { name: 'Dashboard' }
            },
            
            // Dashboard
            {
                path: 'dashboard',
                name: 'Dashboard',
                component: Dashboard,
                meta: { 
                    title: 'Dashboard - G-Backup',
                    icon: '📊'
                }
            },
            
            // Remotes
            {
                path: 'remotes',
                name: 'Remotes',
                component: Remotes,
                meta: { 
                    title: 'Remotes - G-Backup',
                    icon: '☁️'
                }
            },
            
            // Scheduled Jobs
            {
                path: 'scheduled',
                name: 'ScheduledJobs',
                component: ScheduledJobs,
                meta: { 
                    title: 'Scheduled Jobs - G-Backup',
                    icon: '⏰'
                }
            },
            
            // Manual Jobs
            {
                path: 'manual',
                name: 'ManualJobs',
                component: ManualJobs,
                meta: { 
                    title: 'Manual Jobs - G-Backup',
                    icon: '🔧'
                }
            },
            
            // Logs
            {
                path: 'logs',
                name: 'Logs',
                component: Logs,
                meta: { 
                    title: 'Logs - G-Backup',
                    icon: '📝'
                }
            }
        ]
    },
    
    // ----------------------------------------
    // CATCH ALL / 404
    // ----------------------------------------
    {
        path: '/:pathMatch(.*)*',
        name: 'NotFound',
        redirect: { name: 'SetupWizard' } // Default redirect ke login jika 404
    }
];

// ============================================
// CREATE ROUTER INSTANCE
// ============================================
const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes,
    scrollBehavior(to, from, savedPosition) {
        if (savedPosition) {
            return savedPosition;
        } else {
            return { top: 0 };
        }
    }
});

// ============================================
// NAVIGATION GUARDS
// ============================================
router.beforeEach(async (to, from, next) => {
    
    const authStore = useAuthStore();
    const appStore = useAppStore(); // Ambil App Store
    
    // 1. MEMASTIKAN STATUS SETUP TELAH DIMUAT
    if (appStore.setupStatus === 'LOADING') {
        // Tunggu hingga status dimuat dari API
        await appStore.checkSetupStatus(); 
    }

    const appStatus = appStore.setupStatus;
    const isAuthenticated = authStore.isAuthenticated();
    
    // Set document title
    if (to.meta.title) {
        document.title = to.meta.title;
    }
    
    // 2. LOGIKA SETUP WIZARD
    if (appStatus === 'SETUP_NEEDED') {
        // Jika Setup Dibutuhkan, paksa semua rute ke SetupWizard
        if (to.name === 'SetupWizard') {
            next(); // Boleh ke Setup Wizard
        } else {
            console.log('❌ Setup Needed, redirect to Setup Wizard');
            next({ name: 'SetupWizard' }); 
        }
        return;
    }
    
    if (appStatus === 'SETUP_COMPLETE') {
        // Jika Setup Selesai, blokir akses ke rute Setup Wizard
        if (to.meta.isSetupRoute) {
            console.log('❌ Setup Complete, redirect from Setup Wizard');
            // Redirect ke Dashboard jika sudah login, atau Login jika belum
            return next(isAuthenticated ? { name: 'Dashboard' } : { name: 'Login' });
        }
    }

    // 3. LOGIKA OTENTIKASI (Hanya berjalan jika Setup Selesai)
    
    // Check authentication for protected routes
    if (to.meta.requiresAuth && !isAuthenticated) {
        console.log('❌ Not authenticated, redirect to login');
        next({ name: 'Login', query: { redirect: to.fullPath } });
        return;
    }
    
    // Prevent authenticated users from accessing login
    if (to.name === 'Login' && isAuthenticated) {
        console.log('✅ Already authenticated, redirect to dashboard');
        next({ name: 'Dashboard' });
        return;
    }
    
    console.log('✅ Navigation allowed');
    next();
});

// Optional: After navigation
router.afterEach((to, from) => {
    // console.log('✅ Navigated to:', to.path); 
});

export default router;