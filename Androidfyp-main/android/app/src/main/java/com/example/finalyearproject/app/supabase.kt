package com.example.finalyearproject.app.supabase

import io.github.jan.supabase.SupabaseClient
import io.github.jan.supabase.createSupabaseClient
import io.github.jan.supabase.gotrue.Auth
import io.github.jan.supabase.postgrest.Postgrest
import io.github.jan.supabase.storage.Storage

object SupabaseProvider {

    val client: SupabaseClient = createSupabaseClient(
        supabaseUrl = "https://npeusanizvcyjwsgbhfn.supabase.co",
        supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6Im5wZXVzYW5penZjeWp3c2diaGZuIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NTUwODAxMDcsImV4cCI6MjA3MDY1NjEwN30.-l_oZP4XYmD8QDz-MaXNy2AIOr3mSMTdSiBHqQ-k-q8"
    ) {
        install(Auth)
        install(Postgrest)
        install(Storage)
    }
}