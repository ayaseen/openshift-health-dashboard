import React from 'react';

const UploadSection = ({ file, loading, error, handleFileChange, handleUpload }) => {
    return (
        <div className="bg-white shadow rounded-lg p-6 mb-6">
            <h2 className="text-xl font-semibold mb-4">Upload Health Check Report</h2>
            <p className="mb-4 text-gray-600">
                Upload an OpenShift Health Check Report (AsciiDoc format) to visualize the results.
            </p>

            <div className="flex flex-col sm:flex-row space-y-4 sm:space-y-0 sm:space-x-4">
                <div className="flex-1">
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                        Select Report File
                    </label>
                    <div className="flex items-center">
                        <label className="flex-1 cursor-pointer bg-gray-50 px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm text-gray-700 hover:bg-gray-100 focus:outline-none">
                            <span className="truncate">{file ? file.name : 'Choose file...'}</span>
                            <input
                                type="file"
                                accept=".adoc,.asciidoc"
                                onChange={handleFileChange}
                                className="sr-only"
                            />
                        </label>
                        <button
                            onClick={handleUpload}
                            disabled={!file || loading}
                            className={`ml-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white transition-all duration-300
                            ${!file || loading ? 'bg-gray-400 cursor-not-allowed' : 'bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 hover:shadow-md'}`}
                        >
                            {loading ? (
                                <>
                                    <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                    </svg>
                                    Processing...
                                </>
                            ) : (
                                'Upload and Analyze'
                            )}
                        </button>
                    </div>
                    {error && (
                        <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded-md">
                            <p className="text-sm text-red-600">{error}</p>
                        </div>
                    )}
                </div>
            </div>

            {file && !error && !loading && (
                <div className="mt-4 p-3 bg-blue-50 border border-blue-100 rounded-md">
                    <p className="text-sm text-blue-800">
                        <span className="font-medium">Selected file:</span> {file.name} ({(file.size / 1024).toFixed(1)} KB)
                    </p>
                    <p className="text-xs text-blue-700 mt-1">
                        Click "Upload and Analyze" to process the report.
                    </p>
                </div>
            )}

            <div className="mt-4 grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-indigo-50 p-3 rounded-md border border-indigo-100">
                    <h3 className="text-sm font-medium text-indigo-800 mb-1">Dashboard Overview</h3>
                    <p className="text-xs text-indigo-700">View a summary of your cluster health with key metrics</p>
                </div>
                <div className="bg-green-50 p-3 rounded-md border border-green-100">
                    <h3 className="text-sm font-medium text-green-800 mb-1">Executive Summary</h3>
                    <p className="text-xs text-green-700">Generate an executive summary report for stakeholders</p>
                </div>
                <div className="bg-amber-50 p-3 rounded-md border border-amber-100">
                    <h3 className="text-sm font-medium text-amber-800 mb-1">Remediation Steps</h3>
                    <p className="text-xs text-amber-700">Get detailed remediation actions for identified issues</p>
                </div>
            </div>
        </div>
    );
};

export default UploadSection;