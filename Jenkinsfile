pipeline{
    agent any
    
    triggers {
        githubPush()
    }
    
    tools{
        go 'Go'
    }


    stages{
        stage('checkout Github'){
            steps{
                git branch: 'master', credentialsId: 'Jenkins-git', url: 'https://github.com/Lu2ky/04.2-chat-websocket.git'
            }
        }
        stage('Install dependencias'){
            steps{
                sh 'go mod tidy'
            }
        }
        stage('Tests Unitarios'){
            steps{
                sh 'go test -v -cover'
            }
        }
        stage('Test Coverage'){
            steps{
                sh 'go test -coverprofile=coverage.out'
                sh 'go tool cover -func=coverage.out'
            }
        }
        stage('Docker Build'){
            steps{
                script{
                    def image = docker.build("websocket-for-chat:${BUILD_NUMBER}", ".")
                }
            }
        }
        stage('Docker Push'){
            when {
                branch 'master'
            }
            steps{
                withRegistry('https://ghcr.io', 'Jenkins-git'){ 
                    script {
                            docker.image("ghcr.io/tu-usuario/websocket-for-chat:${BUILD_NUMBER}").push()
                            docker.image("ghcr.io/tu-usuario/websocket-for-chat:${BUILD_NUMBER}").push('latest')
                        }
                }
            }
        }
    }
    post{
        success{
            echo 'Build complete'
            mail(
                subject: "BUILD SUCCESS: ${JOB_NAME} #${BUILD_NUMBER}",
                body: """
                    Build exitoso
                    
                    Job: ${JOB_NAME}
                    Build Number: ${BUILD_NUMBER}
                    Build URL: ${BUILD_URL}
                    Status: SUCCESS 
                    
                    Cambios:
                    ${GIT_COMMIT}
                """,
                to: "jaaa736504@gmail.com"
            )
        }
        failure{
            echo 'Build failed'
            mail(
                subject: "❌ BUILD FAILED: ${JOB_NAME} #${BUILD_NUMBER}",
                body: """
                    Build fallo
                    
                    Job: ${JOB_NAME}
                    Build Number: ${BUILD_NUMBER}
                    Build URL: ${BUILD_URL}
                    Status: FAILED ❌
                    
                    Revisa los logs en: ${BUILD_URL}console
                """,
                to: "jaaa736504@gmail.com"
            )
        }
        unstable{
            echo 'Build unstable'
            mail(
                subject: "BUILD UNSTABLE: ${JOB_NAME} #${BUILD_NUMBER}",
                body: """
                    Build inestable!
                    
                    Job: ${JOB_NAME}
                    Build Number: ${BUILD_NUMBER}
                    Build URL: ${BUILD_URL}
                    Status: UNSTABLE 
                """,
                to: "jaaa736504@gmail.com"
            )
        }
    }
}