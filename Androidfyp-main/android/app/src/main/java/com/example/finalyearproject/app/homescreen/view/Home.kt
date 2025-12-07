    package com.example.finalyearproject.app.homescreen.view

    import android.os.Bundle
    import androidx.activity.enableEdgeToEdge
    import androidx.appcompat.app.AppCompatActivity
    import androidx.constraintlayout.widget.ConstraintLayout
    import androidx.core.view.ViewCompat
    import androidx.core.view.WindowInsetsCompat
    import androidx.core.view.updatePadding
    import androidx.fragment.app.Fragment
    import com.example.finalyearproject.R
    import com.example.finalyearproject.app.cvanalysis.view.cvfragment
    import com.example.finalyearproject.app.notifications.view.NotificationFragment
    import com.example.finalyearproject.app.practice.view.Practicefragment
    import com.example.finalyearproject.app.profile.view.ProfileFragment
    import com.google.android.material.bottomnavigation.BottomNavigationView

    class Home : AppCompatActivity() {
        override fun onCreate(savedInstanceState: Bundle?) {
            super.onCreate(savedInstanceState)
            enableEdgeToEdge()
            setContentView(R.layout.activity_home)

            val rootView = findViewById<ConstraintLayout>(R.id.main)
            val bottomNav = findViewById<BottomNavigationView>(R.id.bottom_navigation)

            ViewCompat.setOnApplyWindowInsetsListener(rootView) { view, insets ->
                val systemBars = insets.getInsets(WindowInsetsCompat.Type.systemBars())
                view.setPadding(systemBars.left, systemBars.top, systemBars.right, 0)
                insets
            }

            ViewCompat.setOnApplyWindowInsetsListener(bottomNav) { view, insets ->
                val navBarInsets = insets.getInsets(WindowInsetsCompat.Type.navigationBars())
                view.updatePadding(bottom = navBarInsets.bottom)
                insets
            }

            // Load default fragment only on first launch
            if (savedInstanceState == null) {
                openFragment(Homefragment())
            }

            // Bottom nav listener to switch fragments
            bottomNav.setOnItemSelectedListener { item ->
                when (item.itemId) {
                    R.id.navigation_home -> openFragment(Homefragment())
                    R.id.navigation_practice -> openFragment(Practicefragment())
                    R.id.navigation_cv -> openFragment(cvfragment())
                    R.id.navigation_Notifications -> openFragment(NotificationFragment())
                    R.id.navigation_settings -> openFragment(ProfileFragment())
                    else -> false
                }
            }
        }

        private fun openFragment(fragment: Fragment): Boolean {
            supportFragmentManager.beginTransaction()
                .replace(R.id.fragment_container, fragment)
                .commit()
            return true
        }
    }
