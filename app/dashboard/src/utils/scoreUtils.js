// app/dashboard/src/utils/scoreUtils.js

// Get color based on score value
export const getScoreColor = (score) => {
    if (score >= 90) return '#10B981'; // Green
    if (score >= 70) return '#22C55E'; // Light Green
    if (score >= 50) return '#F59E0B'; // Amber
    return '#EF4444'; // Red
};

// Get rating text based on score value
export const getScoreRating = (score) => {
    if (score >= 90) return 'Excellent';
    if (score >= 70) return 'Good';
    if (score >= 50) return 'Fair';
    return 'Poor';
};

// Get health assessment text based on score value
export const getHealthAssessment = (score) => {
    if (score >= 80) return 'Your cluster is well configured and follows best practices.';
    if (score >= 60) return 'Your cluster is generally healthy but has some areas needing attention.';
    return 'Your cluster needs significant improvements to meet best practices.';
};

// Format score to remove excess decimals
export const formatScore = (score) => {
    const numericScore = parseFloat(score);
    if (isNaN(numericScore)) return '0';

    return Number.isInteger(numericScore)
        ? Math.round(numericScore)
        : parseFloat(numericScore.toFixed(1));
};

// Get category description based on name and score
export const getCategoryDescription = (categoryName, score) => {
    if (score >= 90) {
        return `${categoryName} is excellent with best practices in place.`;
    } else if (score >= 80) {
        return `${categoryName} is well-configured with only minor improvements needed.`;
    } else if (score >= 70) {
        return `${categoryName} meets most requirements but has some areas that could be improved.`;
    } else if (score >= 60) {
        return `${categoryName} has several areas that need attention to meet best practices.`;
    } else {
        return `${categoryName} requires significant improvements to ensure stability and security.`;
    }
};

// Get severity level object with color and label
export const getSeverityLevel = (type) => {
    switch (type.toLowerCase()) {
        case 'required':
            return { color: '#EF4444', label: 'Required', bgColor: '#FEF2F2' };
        case 'recommended':
            return { color: '#F59E0B', label: 'Recommended', bgColor: '#FFFBEB' };
        case 'advisory':
            return { color: '#3B82F6', label: 'Advisory', bgColor: '#EFF6FF' };
        default:
            return { color: '#6B7280', label: 'Info', bgColor: '#F9FAFB' };
    }
};