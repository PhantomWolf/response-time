pipeline {
  agent {
    docker {
      image 'phantomwolf47/golangwithgit:1.0'
    }
    
  }
  stages {
    stage('Build') {
      steps {
        sh 'go get -d ./...'
        sh 'go build'
      }
    }
    stage('Test') {
      steps {
        sh 'go test -v'
      }
    }
  }
  environment {
    CI = 'true'
  }
}
