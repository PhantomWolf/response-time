pipeline {
    agent {
        docker {
            image '1.10-alpine3.7'
        }
    }
    environment {
        CI = 'true'
    }
    stages {
        stage('Build') {
            steps {
                sh 'go build'
            }
        }
        stage('Test') {
            steps {
                sh 'go test'
            }
        }
    }
}


