package com.example.finalyearproject.app.auth.view.adapter

import android.content.Context
import android.widget.ArrayAdapter
import com.example.finalyearproject.R

object SignUpSpinnerAdapter {

    fun getJobPositionAdapter(context: Context): ArrayAdapter<CharSequence> {
        val adapter = ArrayAdapter.createFromResource(
            context,
            R.array.jobpositions,
            android.R.layout.simple_spinner_item
        )
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item)
        return adapter
    }

    fun getSpecializationAdapter(context: Context, major: String): ArrayAdapter<String> {
        val specializations = when (major) {
            "Software Engineering" -> context.resources.getStringArray(R.array.software_engineering_specializations)
            "Business Administration" -> context.resources.getStringArray(R.array.business_administration_specializations)
            "Mechanical Engineering" -> context.resources.getStringArray(R.array.mechanical_engineering_specializations)
            "Graphic Design" -> context.resources.getStringArray(R.array.graphic_design_specializations)
            "Marketing" -> context.resources.getStringArray(R.array.marketing_specializations)
            else -> emptyArray()
        }

        val adapter = ArrayAdapter(
            context,
            android.R.layout.simple_spinner_item,
            specializations.toList()
        )
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item)
        return adapter
    }
}
